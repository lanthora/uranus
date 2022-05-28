// SPDX-License-Identifier: AGPL-3.0-or-later
package control

import (
	"encoding/json"
	"net/http"
	"time"
	"uranus/pkg/connector"

	"github.com/gin-gonic/gin"
)

func Init(engine *gin.Engine) (err error) {
	engine.POST("/control/echo", echo)
	engine.POST("/control/shutdown", shutdown)
	return
}

func echo(context *gin.Context) {
	// 获取前端请求的参数
	request := struct {
		Extra interface{} `json:"extra"`
	}{}

	if err := context.BindJSON(&request); err != nil {
		context.Status(http.StatusBadRequest)
		return
	}

	// 参数组装成底层命令
	requestJson := map[string]interface{}{
		"type":  "user::test::echo",
		"extra": request.Extra,
	}

	bytes, err := json.Marshal(requestJson)
	if err != nil || len(bytes) > 1024 {
		context.Status(http.StatusBadRequest)
		return
	}

	// 转换成字符串向底层发送命令,并接收响应的字符串
	responseStr, err := connector.Exec(string(bytes), time.Second)
	if err != nil {
		context.Status(http.StatusServiceUnavailable)
		return
	}

	// 底层返回的字符串转换后返回给前端
	response := struct {
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		context.Status(http.StatusServiceUnavailable)
		return
	}
	context.JSON(http.StatusOK, response)
}

func shutdown(context *gin.Context) {
}
