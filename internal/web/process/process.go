// SPDX-License-Identifier: AGPL-3.0-or-later
package process

import (
	"database/sql"
	"uranus/internal/config"
	"uranus/internal/web/render"
	"uranus/pkg/process"

	"github.com/gin-gonic/gin"
)

type Worker struct {
	engine *gin.Engine
	db     *sql.DB

	config *config.Config
}

type Event struct {
	ID      uint64 `json:"id"`
	Workdir string `json:"workdir"`
	Binary  string `json:"binary"`
	Argv    string `json:"argv"`
	Count   uint64 `json:"count"`
	Judge   uint64 `json:"judge"`
	Status  uint64 `json:"status"`
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
	w.engine.POST("/process/core/status", w.processCoreStatus)
	w.engine.POST("/process/core/enable", w.processCoreEnable)
	w.engine.POST("/process/core/disable", w.processCoreDisable)
	w.engine.POST("/process/audit/status", w.processAuditStatus)
	w.engine.POST("/process/audit/update", w.processAuditUpdate)
	w.engine.POST("/process/event/list", w.processEventList)
	w.engine.POST("/process/event/delete", w.processEventDelete)
	w.engine.POST("/process/policy/update", w.processPolicyUpdate)
	w.engine.POST("/process/trust/update", w.processTrustUpdate)
	w.engine.POST("/process/trust/status", w.processTrustStatus)
	return
}

func (w *Worker) processCoreStatus(context *gin.Context) {
	status, err := w.config.GetInteger(config.ProcessModuleStatus)
	if err != nil {
		status = process.StatusDisable
	}

	response := struct {
		Status int `json:"status"`
	}{
		Status: status,
	}

	render.Success(context, response)
}
func (w *Worker) processCoreEnable(context *gin.Context) {
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
func (w *Worker) processCoreDisable(context *gin.Context) {
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

func (w *Worker) processAuditStatus(context *gin.Context) {
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

func (w *Worker) processAuditUpdate(context *gin.Context) {
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

func (w *Worker) processEventList(context *gin.Context) {
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

func (w *Worker) processPolicyUpdate(context *gin.Context) {
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

	ok := w.updateStatus(uint64(request.ID), request.Status)
	if !ok {
		render.Status(context, render.StatusProcessUpdatePolicyFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

// TODO: 补充进程事件删除功能
func (w *Worker) processEventDelete(context *gin.Context) {
	render.Status(context, render.StatusUnknownError)
}

func (w *Worker) processTrustUpdate(context *gin.Context) {
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

func (w *Worker) processTrustStatus(context *gin.Context) {
	status, err := w.config.GetInteger(config.ProcessCmdDefaultStatus)
	if err != nil {
		render.Status(context, render.StatusProcessGetTrustStatusFailed)
		return
	}
	response := struct {
		Status int `json:"status" binding:"number"`
	}{
		Status: status,
	}
	render.Success(context, response)
}
