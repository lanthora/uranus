// SPDX-License-Identifier: AGPL-3.0-or-later
package telegram

import (
	"encoding/json"
	"sync"
	"syscall"
	"time"
	"uranus/pkg/connector"

	"github.com/sirupsen/logrus"
)

type TelegramWorker struct {
	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
	bot     *Bot
}

func NewWorker(token string, ownerID int64) *TelegramWorker {
	w := TelegramWorker{
		bot: NewBot(token, ownerID),
	}
	return &w
}

func (w *TelegramWorker) Start() (err error) {
	w.running = true
	err = w.conn.Connect()
	if err != nil {
		return
	}
	err = w.bot.Connect()
	if err != nil {
		return
	}
	err = w.conn.Send(`{"type":"user::msg::sub","section":"audit::proc::report"}`)
	if err != nil {
		return
	}
	w.wg.Add(1)
	go w.runReportToOwner()
	return
}

func (w *TelegramWorker) Stop() (err error) {
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"audit::proc::report"}`)
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	w.running = false
	err = w.conn.Shutdown(time.Now())
	if err != nil {
		return
	}
	w.wg.Wait()
	w.conn.Close()
	return
}

func (w *TelegramWorker) runReportToOwner() {
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

		var doc map[string]interface{}
		err = json.Unmarshal([]byte(msg), &doc)
		if err != nil {
			logrus.Error(err)
			continue
		}

		switch doc["type"].(string) {
		case "audit::proc::report":
			html := RenderAuditProcReport(msg)
			w.bot.SendHtmlToOwner(html)
		case "user::msg::sub":
			html := RenderUserMsgSub(msg)
			w.bot.SendHtmlToOwner(html)
		case "user::msg::unsub":
			html := RenderUserMsgUnsub(msg)
			w.bot.SendHtmlToOwner(html)
		case "kernel::proc::enable":
			html := RenderKernelProcEnable(msg)
			w.bot.SendHtmlToOwner(html)
		case "kernel::proc::disable":
			html := RenderKernelProcDisable(msg)
			w.bot.SendHtmlToOwner(html)
		default:
			w.bot.SendTextToOwner(msg)
		}
	}

}
