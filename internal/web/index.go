// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"github.com/gin-gonic/gin"
)

func index(context *gin.Context) {
	// 这里是 Web 模块在后端项目唯一需要引用的前端资源
	context.File("/usr/share/hackernel/web/index.html")
}
