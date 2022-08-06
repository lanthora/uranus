// SPDX-License-Identifier: AGPL-3.0-or-later
package net

import (
	"database/sql"
	"uranus/internal/config"
	"uranus/internal/web/render"
	"uranus/pkg/net"

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
	w.engine.POST("/net/core/status", w.netCoreStatus)
	w.engine.POST("/net/core/enable", w.netCoreEnable)
	w.engine.POST("/net/core/disable", w.netCoreDisable)
	w.engine.POST("/net/policy/add", w.netPolicyAdd)
	w.engine.POST("/net/policy/delete", w.netPolicyDelete)
	w.engine.POST("/net/policy/list", w.netPolicyList)
	w.engine.POST("/net/event/list", w.netEventList)
	w.engine.POST("/net/event/delete", w.netEventDelete)
	w.engine.POST("/net/event/update", w.netEventUpdate)
	return
}

func (w *Worker) netCoreStatus(context *gin.Context) {
	status, err := w.config.GetInteger(config.NetModuleStatus)
	if err != nil {
		status = net.StatusDisable
	}

	response := struct {
		Status int `json:"status"`
	}{
		Status: status,
	}

	render.Success(context, response)
}

func (w *Worker) netCoreEnable(context *gin.Context) {
	if err := w.config.SetInteger(config.NetModuleStatus, net.StatusEnable); err != nil {
		render.Status(context, render.StatusProcessEnableFailed)
		return
	}
	ok := net.Enable()
	if !ok {
		render.Status(context, render.StatusNetEnableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) netCoreDisable(context *gin.Context) {
	if err := w.config.SetInteger(config.NetModuleStatus, net.StatusDisable); err != nil {
		render.Status(context, render.StatusNetDisableFailed)
		return
	}
	ok := net.Disable()
	if !ok {
		render.Status(context, render.StatusNetDisableFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) netPolicyAdd(context *gin.Context) {
	request := net.Policy{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	id, err := w.insertNetPolicy(&request)
	if err != nil {
		render.Status(context, render.StatusNetAddPolicyDatabaseFailed)
		return
	}

	request.ID = id
	if ok := net.AddPolicy(request); !ok {
		render.Status(context, render.StatusNetAddPolicyFailed)
		return
	}

	render.Status(context, render.StatusSuccess)
}

func (w *Worker) netPolicyDelete(context *gin.Context) {
	request := struct {
		ID int `json:"id" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	err := w.deleteNetPolicyById(request.ID)
	if err == net.ErrorPolicyNotExist {
		render.Status(context, render.StatusNetPolicyNotExist)
		return
	}

	if err != nil {
		render.Status(context, render.StatusNetDeletePolicyDatabaseFailed)
		return
	}

	ok := net.DeletePolicy(request.ID)
	if !ok {
		render.Status(context, render.StatusNetDeletePolicyFailed)
		return
	}

	render.Status(context, render.StatusSuccess)
}

func (w *Worker) netPolicyList(context *gin.Context) {
	request := struct {
		Limit  int `json:"limit" binding:"number"`
		Offset int `json:"offset" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	policies, err := w.queryNetPolicyLimitOffset(request.Limit, request.Offset)
	if err != nil {
		logrus.Error(err)
		render.Status(context, render.StatusNetQueryPolicyListFailed)
		return
	}
	render.Success(context, policies)
}

func (w *Worker) netEventList(context *gin.Context) {
	request := struct {
		Limit  int `json:"limit" binding:"number"`
		Offset int `json:"offset" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	events, err := w.queryNetEventOffsetLimit(request.Limit, request.Offset)
	if err != nil {
		render.Status(context, render.StatusNetQueryEventFailed)
		return
	}
	render.Success(context, events)
}

func (w *Worker) netEventDelete(context *gin.Context) {
	request := struct {
		ID int `json:"id" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	err := w.deleteNetEventById(request.ID)
	if err != nil {
		render.Status(context, render.StatusNetDeleteEventFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) netEventUpdate(context *gin.Context) {
	request := struct {
		Status int `json:"status" binding:"number"`
		ID     int `json:"id" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	err := w.updateNetEventStatusById(request.Status, request.ID)
	if err != nil {
		render.Status(context, render.StatusNetUpdateEventStatusFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}
