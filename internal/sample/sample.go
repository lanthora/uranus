package sample

import (
	"sync"
	"time"
	"uranus/pkg/connector"

	"github.com/sirupsen/logrus"
)

type SampleWorker struct {
	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
}

func (w *SampleWorker) Start() {
	logrus.Debug("Start")
	w.running = true
	w.conn.Connect()
	w.conn.Send(`{"type":"user::proc::enable"}`)
	w.conn.Send(`{"type":"user::msg::sub","section":"kernel::proc::report"}`)
	w.wg.Add(1)
	go w.run()
}

func (w *SampleWorker) run() {
	for w.running {
		msg, err := w.conn.Recv()
		if err == nil {
			logrus.Debugf("msg=[%s]", msg)
		}
	}
	w.wg.Done()
}

func (w *SampleWorker) Stop() {
	w.conn.Send(`{"type":"user::proc::disable"}`)
	w.conn.Send(`{"type":"user::msg::unsub","section":"kernel::proc::report"}`)
	time.Sleep(time.Millisecond * 100)
	w.running = false
	w.conn.Shutdown()
	w.wg.Wait()
	w.conn.Close()
	logrus.Debug("Stop")
}
