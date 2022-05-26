// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"embed"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed static
var staticFS embed.FS

var contentType = map[string]string{
	".html": "text/html; charset=UTF-8",
	".css":  "text/css; charset=UTF-8",
	".js":   "text/javascript; charset=UTF-8",
}

func static(context *gin.Context) {
	filepath := strings.TrimPrefix(context.Request.URL.String(), "/")
	if filepath == "" {
		filepath = filepath + "index"
	}
	if !strings.HasPrefix(filepath, "static/") {
		filepath = "static/" + filepath
	}
	if path.Ext(filepath) == "" {
		filepath = filepath + ".html"
	}
	if data, err := staticFS.ReadFile(filepath); err == nil {
		context.Data(http.StatusOK, contentType[path.Ext(filepath)], data)
	} else {
		context.Status(http.StatusNotFound)
	}
}
