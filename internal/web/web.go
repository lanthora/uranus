// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"syscall"
	"time"

	"uranus/internal/web/control"
	"uranus/internal/web/file"
	"uranus/internal/web/net"
	"uranus/internal/web/process"
	"uranus/internal/web/user"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type WebWorker struct {
	addr   string
	server *http.Server
	wg     sync.WaitGroup
	db     *sql.DB
}

func NewWorker(addr string, db *sql.DB) *WebWorker {
	w := WebWorker{
		addr: addr,
		db:   db,
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

func (w *WebWorker) Init() (err error) {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.GET("/*filename", front)

	if err = user.Init(engine, w.db); err != nil {
		return
	}

	if err = process.Init(engine, w.db); err != nil {
		return
	}

	if err = file.Init(engine, w.db); err != nil {
		return
	}

	if err = net.Init(engine, w.db); err != nil {
		return
	}

	control.Init(engine, w.db)

	w.server = &http.Server{
		Addr:    w.addr,
		Handler: engine,
	}
	return
}

func (w *WebWorker) Start() (err error) {
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
