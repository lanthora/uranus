// SPDX-License-Identifier: AGPL-3.0-or-later
package ctrl

import (
	"database/sql"
	"strings"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/lanthora/uranus/internal/web/render"
	"github.com/lanthora/uranus/pkg/ctrl"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	db           *sql.DB
	debugEnabled bool
}

func Init(engine *gin.Engine, db *sql.DB) (err error) {
	w := &Worker{
		db:           db,
		debugEnabled: false,
	}

	engine.Use(w.middleware())
	engine.POST("/ctrl/shutdown", w.shutdown)
	engine.GET("/ctrl/enableDebug", w.enableDebug)
	engine.GET("/ctrl/disableDebug", w.disableDebug)
	return
}

func (w *Worker) shutdown(context *gin.Context) {
	ctrl.Shutdown()
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) enableDebug(context *gin.Context) {
	w.debugEnabled = true
	logrus.Info("pprof debug enabled")
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) disableDebug(context *gin.Context) {
	w.debugEnabled = false
	logrus.Info("pprof debug disabled")
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) middleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		if strings.HasPrefix(context.Request.URL.Path, pprof.DefaultPrefix) {
			if !w.debugEnabled {
				context.Abort()
				return
			}
		}
		context.Next()
	}
}
