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
}

func NewWorker(addr string) *WebWorker {
	w := WebWorker{
		addr: addr,
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

	engine.Use()

	engine.GET("/", static)
	engine.GET("/favicon.ico", static)
	engine.GET("/static/*filename", static)
	engine.GET("/login", static)

	control.RegisterRoute(engine)
	user.RegisterRoute(engine)

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
