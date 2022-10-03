// SPDX-License-Identifier: AGPL-3.0-or-later
package ctrl

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/uranus/internal/web/render"
	"github.com/lanthora/uranus/pkg/ctrl"
)

type Worker struct {
	db *sql.DB
}

func Init(engine *gin.Engine, db *sql.DB) (err error) {
	w := &Worker{
		db: db,
	}

	engine.POST("/ctrl/shutdown", w.shutdown)
	return
}

func (w *Worker) shutdown(context *gin.Context) {
	ctrl.Shutdown()
	render.Status(context, render.StatusSuccess)
}
