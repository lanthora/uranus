// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"embed"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

//go:embed webui/*
var staticFS embed.FS

var contentType = map[string]string{
	".html": "text/html; charset=UTF-8",
	".css":  "text/css; charset=UTF-8",
	".js":   "text/javascript; charset=UTF-8",
}

func front(context *gin.Context) {
	filepath := context.Request.URL.String()
	if filepath == "/" {
		filepath = filepath + "index.html"
	}
	filepath = "webui" + filepath
	if data, err := staticFS.ReadFile(filepath); err == nil {
		context.Data(http.StatusOK, contentType[path.Ext(filepath)], data)
	} else {
		context.Status(http.StatusNotFound)
	}
}
