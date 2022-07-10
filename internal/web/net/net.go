// SPDX-License-Identifier: AGPL-3.0-or-later
package net

import (
	"uranus/internal/config"
	"uranus/internal/web/render"
	"uranus/pkg/net"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	engine         *gin.Engine
	dataSourceName string

	config *config.Config
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
	w.engine.POST("/net/core/status", w.netCoreStatus)
	w.engine.POST("/net/core/enable", w.netCoreEnable)
	w.engine.POST("/net/core/disable", w.netCoreDisable)
	w.engine.POST("/net/policy/add", w.netPolicyAdd)
	w.engine.POST("/net/policy/delete", w.netPolicyDelete)
	w.engine.POST("/net/policy/list", w.netPolicyList)
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

	if ok := net.AddPolicy(request); !ok {
		render.Status(context, render.StatusAddNetPolicyFailed)
		return
	}

	err := w.insertNetPolicy(&request)
	if err != nil {
		render.Status(context, render.StatusAddNetPolicyDatabaseFailed)
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

	ok := net.DeletePolicy(request.ID)
	if !ok {
		render.Status(context, render.StatusDeleteNetPolicyFailed)
		return
	}

	err := w.deleteNetPolicyById(request.ID)
	if err != nil {
		render.Status(context, render.StatusDeleteNetPolicyDatabaseFailed)
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
		render.Status(context, render.StatusQueryNetPolicyListFailed)
		return
	}
	render.Success(context, policies)
}
