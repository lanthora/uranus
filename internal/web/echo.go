// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"encoding/json"
	"net/http"
	"time"
	"uranus/pkg/connector"

	"github.com/gin-gonic/gin"
)

func echo(context *gin.Context) {
	requestJson := map[string]string{
		"type":  "user::test::echo",
		"extra": context.DefaultQuery("extra", ""),
	}

	requestBytes, err := json.Marshal(requestJson)
	if err != nil {
		context.Status(http.StatusServiceUnavailable)
		return
	}

	response, err := connector.Exec(string(requestBytes), time.Second)
	if err != nil {
		context.Status(http.StatusServiceUnavailable)
		return
	}
	context.String(http.StatusOK, response)
}
