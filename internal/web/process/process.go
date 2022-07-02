// SPDX-License-Identifier: AGPL-3.0-or-later
package process

import (
	"uranus/internal/config"
	"uranus/internal/web/render"
	"uranus/pkg/process"

	"github.com/gin-gonic/gin"
)

type Worker struct {
	engine         *gin.Engine
	dataSourceName string

	config *config.Config
}

type Event struct {
	ID      uint64 `json:"id" binding:"required"`
	Workdir string `json:"workdir" binding:"required"`
	Binary  string `json:"binary" binding:"required"`
	Argv    string `json:"argv" binding:"required"`
	Count   uint64 `json:"count" binding:"required"`
	Judge   uint64 `json:"judge" binding:"required"`
	Status  uint64 `json:"status" binding:"required"`
}

func Init(engine *gin.Engine, dataSourceName string) (err error) {
	config, err := config.New(dataSourceName)
	if err != nil {
		return
	}
	w := &Worker{
		engine:         engine,
		dataSourceName: dataSourceName,
		config:         config,
	}
	w.engine.POST("/process/core/status", w.processCoreStatus)
	w.engine.POST("/process/core/enable", w.processCoreEnable)
	w.engine.POST("/process/core/disable", w.processCoreDisable)
	w.engine.POST("/process/judge/status", w.processJudgeStatus)
	w.engine.POST("/process/judge/update", w.processJudgeUpdate)
	w.engine.POST("/process/event/list", w.processEventList)
	w.engine.POST("/process/event/update", w.processEventUpdate)
	w.engine.POST("/process/event/delete", w.processEventDelete)
	w.engine.POST("/process/auto/trust", w.processAutoTrust)
	w.engine.POST("/process/auto/status", w.processAutoStatus)
	return
}

func (w *Worker) processCoreStatus(context *gin.Context) {
	status, err := w.config.GetInteger("proc::core::status")
	if err != nil {
		status = process.StatusDisable
	}

	response := struct {
		Core int `json:"core" binding:"required"`
	}{
		Core: status,
	}

	render.Success(context, response)
}
func (w *Worker) processCoreEnable(context *gin.Context) {
	if err := w.config.SetInteger("proc::core::status", process.StatusEnable); err != nil {
		render.Status(context, render.StatusProcessEnableFailed)
		return
	}
	ok := process.ProcessEnable()
	if !ok {
		render.Status(context, render.StatusProcessEnableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}
func (w *Worker) processCoreDisable(context *gin.Context) {
	if err := w.config.SetInteger("proc::core::status", process.StatusDisable); err != nil {
		render.Status(context, render.StatusProcessDisableFailed)
		return
	}
	ok := process.ProcessDisable()
	if !ok {
		render.Status(context, render.StatusProcessDisableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) processJudgeStatus(context *gin.Context) {
	judge, err := w.config.GetInteger("proc::judge::status")
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

func (w *Worker) processJudgeUpdate(context *gin.Context) {
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

	if err := w.config.SetInteger("proc::judge::status", request.Judge); err != nil {
		render.Status(context, render.StatusUpdateProcessJudgeFailed)
		return
	}
	ok := process.ProcessJudgeUpdate(request.Judge)
	if !ok {
		render.Status(context, render.StatusUpdateProcessJudgeFailed)
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
		render.Status(context, render.StatusQueryProcessEventFailed)
		return
	}
	render.Success(context, events)
}

func (w *Worker) processEventUpdate(context *gin.Context) {
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
		render.Status(context, render.StatusUpdateProcessEventStatusFailed)
		return
	}

	ok := w.updateStatus(uint64(request.ID), request.Status)
	if !ok {
		render.Status(context, render.StatusUpdateProcessEventStatusFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

// TODO: 补充进程事件删除功能
func (w *Worker) processEventDelete(context *gin.Context) {
	render.Status(context, render.StatusUnknownError)
}

func (w *Worker) processAutoTrust(context *gin.Context) {
	request := struct {
		Status int `json:"status" binding:"number"`
	}{}
	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}
	if err := w.config.SetInteger("proc::auto::trust", request.Status); err != nil {
		render.Status(context, render.StatusProcessAutoTrustFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) processAutoStatus(context *gin.Context) {
	status, err := w.config.GetInteger("proc::auto::trust")
	if err != nil {
		render.Status(context, render.StatusProcessGetAutoStatusFailed)
		return
	}
	response := struct {
		Status int `json:"status" binding:"number"`
	}{
		Status: status,
	}
	render.Success(context, response)
}
