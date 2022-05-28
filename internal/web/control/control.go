// SPDX-License-Identifier: AGPL-3.0-or-later
package control

import (
	"encoding/json"
	"net/http"
	"time"
	"uranus/pkg/connector"

	"github.com/gin-gonic/gin"
)

func RegisterRoute(engine *gin.Engine) {
	engine.POST("/control/echo", echo)
	engine.POST("/control/shutdown", shutdown)
}

func echo(context *gin.Context) {
	request := struct {
		Session string `json:"session" binding:"required"`
		Extra   string `json:"extra" binding:"required"`
	}{}

	if err := context.BindJSON(&request); err != nil {
		context.Status(http.StatusBadRequest)
		return
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}

	response, err := connector.Exec(string(bytes), time.Second)
	if err != nil {
		context.Status(http.StatusServiceUnavailable)
		return
	}
	context.String(http.StatusOK, response)
}

func shutdown(context *gin.Context) {
}
