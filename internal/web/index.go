// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed index.html
	indexContent []byte
)

func index(context *gin.Context) {
	// 这里是 Web 模块在后端项目唯一需要引用的前端资源
	context.Data(http.StatusOK, "text/html; charset=utf-8", indexContent)
}
