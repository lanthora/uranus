// SPDX-License-Identifier: AGPL-3.0-or-later
package control

import (
	"encoding/json"
	"time"
	"uranus/internal/web/render"
	"uranus/pkg/connector"

	"github.com/gin-gonic/gin"
)

func Init(engine *gin.Engine, dataSourceName string) (err error) {
	engine.POST("/control/echo", echo)
	engine.POST("/control/shutdown", shutdown)
	return
}

func echo(context *gin.Context) {
	// 获取前端请求的参数
	request := struct {
		Extra interface{} `json:"extra"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	// 参数组装成底层命令
	requestJson := map[string]interface{}{
		"type":  "user::test::echo",
		"extra": request.Extra,
	}

	bytes, err := json.Marshal(requestJson)
	if err != nil || len(bytes) > 1024 {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	// 转换成字符串向底层发送命令,并接收响应的字符串
	responseStr, err := connector.Exec(string(bytes), time.Second)
	if err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}

	// 底层返回的字符串转换后返回给前端
	response := struct {
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}
	render.Success(context, response)
}

func shutdown(context *gin.Context) {
}
