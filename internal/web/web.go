// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"context"
	"net/http"
	"sync"
	"syscall"
	"time"

	"uranus/internal/web/control"
	"uranus/internal/web/user"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type WebWorker struct {
	addr   string
	server *http.Server
	wg     sync.WaitGroup
	dbName string
}

func NewWorker(addr string, dbName string) *WebWorker {
	w := WebWorker{
		addr:   addr,
		dbName: dbName,
	}
	return &w
}

func (w *WebWorker) serve() {
	defer w.wg.Done()
	if err := w.server.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Error(err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}
}

func (w *WebWorker) Start() (err error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(user.Middleware())

	engine.GET("/*filename", front)
	control.Init(engine, w.dbName)
	user.Init(engine, w.dbName)

	w.server = &http.Server{
		Addr:    w.addr,
		Handler: engine,
	}
	w.wg.Add(1)
	go w.serve()
	return
}

func (w *WebWorker) Stop() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.server.Shutdown(ctx)
	w.wg.Wait()
	return
}
