// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"embed"
	"net/http"
	"path"

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
	filename, ok := context.Params.Get("filename")
	if !ok {
		filename = "/index.html"
	}
	filename = "static" + filename
	if data, err := staticFS.ReadFile(filename); err == nil {
		context.Data(http.StatusOK, contentType[path.Ext(filename)], data)
	} else {
		context.Status(http.StatusNotFound)
	}

}
