// SPDX-License-Identifier: AGPL-3.0-or-later
package file

import (
	"database/sql"
	"uranus/internal/config"
	"uranus/internal/web/render"
	"uranus/pkg/file"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	engine *gin.Engine
	db     *sql.DB

	config *config.Config
}

func Init(engine *gin.Engine, db *sql.DB) (err error) {
	config, err := config.New(db)
	if err != nil {
		return
	}
	w := &Worker{
		engine: engine,
		db:     db,
		config: config,
	}
	w.engine.POST("/file/core/status", w.fileCoreStatus)
	w.engine.POST("/file/core/enable", w.fileCoreEnable)
	w.engine.POST("/file/core/disable", w.fileCoreDisable)
	w.engine.POST("/file/policy/add", w.filePolicyAdd)
	w.engine.POST("/file/policy/update", w.filePolicyUpdate)
	w.engine.POST("/file/policy/delete", w.filePolicyDelete)
	w.engine.POST("/file/policy/clear", w.filePolicyClear)
	w.engine.POST("/file/policy/list", w.filePolicyList)
	w.engine.POST("/file/policy/query", w.filePolicyQuery)
	w.engine.POST("/file/event/list", w.fileEventList)
	w.engine.POST("/file/event/delete", w.fileEventDelete)
	w.engine.POST("/file/event/update", w.fileEventUpdate)
	return
}

func (w *Worker) fileCoreStatus(context *gin.Context) {
	status, err := w.config.GetInteger(config.FileModuleStatus)
	if err != nil {
		status = file.StatusDisable
	}

	response := struct {
		Status int `json:"status"`
	}{
		Status: status,
	}

	render.Success(context, response)
}

func (w *Worker) fileCoreEnable(context *gin.Context) {
	if err := w.config.SetInteger(config.FileModuleStatus, file.StatusEnable); err != nil {
		render.Status(context, render.StatusProcessEnableFailed)
		return
	}
	ok := file.Enable()
	if !ok {
		render.Status(context, render.StatusProcessEnableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) fileCoreDisable(context *gin.Context) {
	if err := w.config.SetInteger(config.FileModuleStatus, file.StatusDisable); err != nil {
		render.Status(context, render.StatusFileDisableFailed)
		return
	}
	ok := file.Disable()
	if !ok {
		render.Status(context, render.StatusFileDisableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) filePolicyAdd(context *gin.Context) {
	request := struct {
		Path string `json:"path" binding:"required"`
		Perm int    `json:"perm" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}
	fsid, ino, status, err := file.SetPolicy(request.Path, request.Perm, file.FlagNew)
	if err != nil || status == file.StatusPolicyUnknown {
		render.Status(context, render.StatusUnknownError)
		return
	}

	if status == file.StatusPolicyConflict {
		render.Status(context, render.StatusFileAddPolicyConflict)
		return
	}

	if status == file.StatusPolicyConflict {
		render.Status(context, render.StatusFileAddPolicyFileNotExist)
		return
	}

	err = w.insertFilePolicy(request.Path, fsid, ino, request.Perm, file.StatusPolicyNormal)
	if err != nil {
		render.Status(context, render.StatusFileAddPolicyFailed)
		return
	}

	render.Status(context, render.StatusSuccess)
}

func (w *Worker) filePolicyUpdate(context *gin.Context) {
	request := struct {
		ID   int `json:"id" binding:"number"`
		Perm int `json:"perm" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	policy, err := w.queryFilePolicyById(request.ID)
	if err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}

	fsid, ino, status, err := file.SetPolicy(policy.Path, request.Perm, file.FlagUpdate)
	if err != nil || status == file.StatusPolicyUnknown {
		render.Status(context, render.StatusUnknownError)
		return
	}

	if status == file.StatusPolicyConflict {
		render.Status(context, render.StatusFileUpdatePolicyConflict)
		return
	}

	if status == file.StatusPolicyConflict {
		render.Status(context, render.StatusFileUpdatePolicyFileNotExist)
		return
	}

	err = w.updateFilePolicyById(fsid, ino, request.Perm, file.StatusPolicyNormal, request.ID)
	if err != nil {
		logrus.Error(err)
		render.Status(context, render.StatusFileUpdatePolicyFailed)
		return
	}

	render.Status(context, render.StatusSuccess)
}

func (w *Worker) filePolicyDelete(context *gin.Context) {
	request := struct {
		ID int `json:"id" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	policy, err := w.queryFilePolicyById(request.ID)
	if err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}

	_, _, _, err = file.SetPolicy(policy.Path, 0, file.FlagAny)
	if err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}

	err = w.deleteFilePolicyById(request.ID)
	if err != nil {
		render.Status(context, render.StatusFileDeletePolicyFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

// TODO: 清空配置列表
func (w *Worker) filePolicyClear(context *gin.Context) {
	render.Status(context, render.StatusUnknownError)
}

func (w *Worker) filePolicyList(context *gin.Context) {
	request := struct {
		Limit  int `json:"limit" binding:"number"`
		Offset int `json:"offset" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	policies, err := w.queryFilePolicyLimitOffset(request.Limit, request.Offset)
	if err != nil {
		logrus.Error(err)
		render.Status(context, render.StatusFileQueryPolicyListFailed)
		return
	}
	render.Success(context, policies)
}

func (w *Worker) filePolicyQuery(context *gin.Context) {
	request := struct {
		ID int `json:"id" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	policy, err := w.queryFilePolicyById(request.ID)
	if err != nil {
		render.Status(context, render.StatusFileQueryPolicyByIdFailed)
		return
	}
	render.Success(context, policy)
}

func (w *Worker) fileEventList(context *gin.Context) {
	request := struct {
		Limit  int `json:"limit" binding:"number"`
		Offset int `json:"offset" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	events, err := w.queryFileEventOffsetLimit(request.Limit, request.Offset)
	if err != nil {
		render.Status(context, render.StatusProcessQueryEventFailed)
		return
	}
	render.Success(context, events)
}

func (w *Worker) fileEventDelete(context *gin.Context) {
	request := struct {
		ID int `json:"id" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	err := w.deleteFileEventById(request.ID)
	if err != nil {
		render.Status(context, render.StatusFileDeleteEventFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) fileEventUpdate(context *gin.Context) {
	request := struct {
		Status int `json:"status" binding:"number"`
		ID     int `json:"id" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	err := w.updateFileEventStatusById(request.Status, request.ID)
	if err != nil {
		render.Status(context, render.StatusFileUpdateEventStatusFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}
