// SPDX-License-Identifier: AGPL-3.0-or-later
package sample

import (
	"sync"
	"syscall"
	"time"
	"uranus/pkg/connector"

	"github.com/sirupsen/logrus"
)

type SampleWorker struct {
	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
}

func NewWorker() *SampleWorker {
	w := SampleWorker{}
	return &w
}

func (w *SampleWorker) Start() {
	logrus.Debug("Start")
	w.running = true
	err := w.conn.Connect()
	if err != nil {
		logrus.Fatal(err)
	}
	err = w.conn.Send(`{"type":"user::proc::enable"}`)
	if err != nil {
		logrus.Fatal(err)
	}
	err = w.conn.Send(`{"type":"user::msg::sub","section":"kernel::proc::report"}`)
	if err != nil {
		logrus.Fatal(err)
	}
	w.wg.Add(1)
	go w.run()
}

func (w *SampleWorker) run() {
	defer w.wg.Done()
	for w.running {
		msg, err := w.conn.Recv()

		if !w.running {
			break
		}

		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			continue
		}

		logrus.Debugf("msg=[%s]", msg)
	}
}

func (w *SampleWorker) Stop() {
	w.conn.Send(`{"type":"user::proc::disable"}`)
	w.conn.Send(`{"type":"user::msg::unsub","section":"kernel::proc::report"}`)
	time.Sleep(time.Second)
	w.running = false
	err := w.conn.Shutdown()
	if err != nil {
		logrus.Fatal(err)
	}
	w.wg.Wait()
	w.conn.Close()
	logrus.Debug("Stop")
}
