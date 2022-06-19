// SPDX-License-Identifier: AGPL-3.0-or-later
package process

import (
	"encoding/json"
	"time"
	"uranus/internal/config"
	"uranus/internal/web/render"
	"uranus/pkg/connector"
	"uranus/pkg/process"
	"uranus/pkg/status"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	engine *gin.Engine
	dbName string

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

func Init(engine *gin.Engine, dbName string) (err error) {
	config, err := config.New(dbName)
	if err != nil {
		return
	}
	w := &Worker{
		engine: engine,
		dbName: dbName,
		config: config,
	}
	w.engine.POST("/process/core/status", w.coreStatus)
	w.engine.POST("/process/core/enable", w.coreEnable)
	w.engine.POST("/process/core/disable", w.coreDisable)
	w.engine.POST("/process/judge/status", w.judgeStatus)
	w.engine.POST("/process/judge/update", w.judgeUpdate)
	w.engine.POST("/process/event/list", w.eventList)
	w.engine.POST("/process/event/update", w.eventUpdate)
	return
}

func (w *Worker) coreStatus(context *gin.Context) {
	core, err := w.config.GetInteger("proc::core::status")
	if err != nil {
		core = status.ProcessCoreDisable
	}

	response := struct {
		Core int `json:"core" binding:"required"`
	}{
		Core: core,
	}

	render.Success(context, response)
}
func (w *Worker) coreEnable(context *gin.Context) {
	if err := w.config.SetInteger("proc::core::status", status.ProcessCoreEnable); err != nil {
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
func (w *Worker) coreDisable(context *gin.Context) {
	if err := w.config.SetInteger("proc::core::status", status.ProcessCoreDisable); err != nil {
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

func (w *Worker) judgeStatus(context *gin.Context) {
	judge, err := w.config.GetInteger("proc::judge::status")
	if err != nil {
		judge = status.ProcessJudgeDisable
	}

	response := struct {
		Judge int `json:"judge" binding:"number"`
	}{
		Judge: judge,
	}

	render.Success(context, response)
}

func (w *Worker) judgeUpdate(context *gin.Context) {
	request := struct {
		Judge int `json:"judge" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if request.Judge < status.ProcessJudgeDisable || request.Judge > status.ProcessJudgeProtect {
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

func (w *Worker) eventList(context *gin.Context) {
	request := struct {
		Limit  *int `json:"limit" binding:"number"`
		Offset *int `json:"offset" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	events, err := w.queryLimitOffset(*request.Limit, *request.Offset)
	if err != nil {
		render.Status(context, render.StatusQueryProcessEventFailed)
		return
	}
	render.Success(context, events)
}

func setTrustedCmd(cmd string) (err error) {
	data := map[string]string{
		"type": "user::proc::trusted::insert",
		"cmd":  cmd,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}

	// TODO: 更细致的判断是否执行成功
	_, err = connector.Exec(string(b), time.Second)
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func (w *Worker) eventUpdate(context *gin.Context) {
	request := struct {
		ID     int `json:"id" binding:"number"`
		Status int `json:"status" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if request.Status == status.ProcessTrusted {
		cmd, err := w.queryCmdById(request.ID)
		if err != nil {
			render.Status(context, render.StatusInvalidArgument)
			return
		}
		err = setTrustedCmd(cmd)
		if err != nil {
			render.Status(context, render.StatusUpdateProcessEventStatusFailed)
			return
		}
	}

	ok := w.updateStatus(uint64(request.ID), request.Status)
	if !ok {
		render.Status(context, render.StatusUpdateProcessEventStatusFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}
