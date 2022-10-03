// SPDX-License-Identifier: AGPL-3.0-or-later
package process

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/uranus/internal/config"
	"github.com/lanthora/uranus/internal/web/render"
	"github.com/lanthora/uranus/pkg/process"
)

type Worker struct {
	db *sql.DB

	config *config.Config
}

type Event struct {
	ID      int64  `json:"id"`
	Workdir string `json:"workdir"`
	Binary  string `json:"binary"`
	Argv    string `json:"argv"`
	Count   int64  `json:"count"`
	Judge   int64  `json:"judge"`
	Status  int64  `json:"status"`
}

func Init(engine *gin.Engine, db *sql.DB) (err error) {
	config, err := config.New(db)
	if err != nil {
		return
	}
	w := &Worker{
		db:     db,
		config: config,
	}
	engine.POST("/process/enableModule", w.enableModule)
	engine.POST("/process/disableModule", w.disableModule)
	engine.POST("/process/showModuleStatus", w.showModuleStatus)

	engine.POST("/process/updateWorkMode", w.updateWorkMode)
	engine.POST("/process/showWorkMode", w.showWorkMode)

	engine.POST("/process/updateEventStatus", w.updateEventStatus)
	engine.POST("/process/deleteEvents", w.deleteEvents)
	engine.POST("/process/listEvents", w.listEvents)

	engine.POST("/process/updateDefaultEventStatus", w.updateDefaultEventStatus)
	engine.POST("/process/showDefaultEventStatus", w.showDefaultEventStatus)
	return
}

func (w *Worker) showModuleStatus(context *gin.Context) {
	status, err := w.config.GetInteger(config.ProcessModuleStatus)
	if err != nil {
		status = process.StatusDisable
	}

	policyCount, err := w.queryProcessPolicyCount()
	if err != nil {
		policyCount = 0
	}
	unreadEventCount, err := w.queryProcessUnreadEventCount()
	if err != nil {
		unreadEventCount = 0
	}

	response := struct {
		Status           int `json:"status"`
		PolicyCount      int `json:"policyCount"`
		UnreadEventCount int `json:"unreadEventCount"`
	}{
		Status:           status,
		PolicyCount:      policyCount,
		UnreadEventCount: unreadEventCount,
	}
	render.Success(context, response)
}
func (w *Worker) enableModule(context *gin.Context) {
	if err := w.config.SetInteger(config.ProcessModuleStatus, process.StatusEnable); err != nil {
		render.Status(context, render.StatusProcessEnableFailed)
		return
	}
	ok := process.Enable()
	if !ok {
		render.Status(context, render.StatusProcessEnableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}
func (w *Worker) disableModule(context *gin.Context) {
	if err := w.config.SetInteger(config.ProcessModuleStatus, process.StatusDisable); err != nil {
		render.Status(context, render.StatusProcessDisableFailed)
		return
	}
	ok := process.Disable()
	if !ok {
		render.Status(context, render.StatusProcessDisableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) showWorkMode(context *gin.Context) {
	judge, err := w.config.GetInteger(config.ProcessProtectionMode)
	if err != nil {
		judge = process.StatusJudgeDisable
	}

	response := struct {
		Judge int `json:"judge" binding:"number"`
	}{
		Judge: judge,
	}

	render.Success(context, response)
}

func (w *Worker) updateWorkMode(context *gin.Context) {
	request := struct {
		Judge int `json:"judge" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if request.Judge < process.StatusJudgeDisable || request.Judge > process.StatusJudgeDefense {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if err := w.config.SetInteger(config.ProcessProtectionMode, request.Judge); err != nil {
		render.Status(context, render.StatusProcessUpdateJudgeFailed)
		return
	}
	ok := process.UpdateJudge(request.Judge)
	if !ok {
		render.Status(context, render.StatusProcessUpdateJudgeFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) listEvents(context *gin.Context) {
	request := struct {
		Limit  int `json:"limit" binding:"number"`
		Offset int `json:"offset" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	events, err := w.queryLimitOffset(request.Limit, request.Offset)
	if err != nil {
		render.Status(context, render.StatusProcessQueryEventFailed)
		return
	}
	render.Success(context, events)
}

func (w *Worker) updateEventStatus(context *gin.Context) {
	request := struct {
		ID     int `json:"id" binding:"number"`
		Status int `json:"status" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	cmd, err := w.queryCmdById(request.ID)
	if err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	switch request.Status {
	case process.StatusTrusted:
		err = process.SetTrustedCmd(cmd)
	default:
		err = process.SetUntrustedCmd(cmd)
	}

	if err != nil {
		render.Status(context, render.StatusProcessUpdatePolicyFailed)
		return
	}

	ok := w.updateStatus(int64(request.ID), request.Status)
	if !ok {
		render.Status(context, render.StatusProcessUpdatePolicyFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

// TODO: 补充进程事件删除功能
func (w *Worker) deleteEvents(context *gin.Context) {
	render.Status(context, render.StatusUnknownError)
}

func (w *Worker) updateDefaultEventStatus(context *gin.Context) {
	request := struct {
		Status int `json:"status" binding:"number"`
	}{}
	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}
	if err := w.config.SetInteger(config.ProcessCmdDefaultStatus, request.Status); err != nil {
		render.Status(context, render.StatusProcessTrustUpdateFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) showDefaultEventStatus(context *gin.Context) {
	status, err := w.config.GetInteger(config.ProcessCmdDefaultStatus)
	if err != nil {
		status = process.StatusPending
	}
	response := struct {
		Status int `json:"status" binding:"number"`
	}{
		Status: status,
	}
	render.Success(context, response)
}
