package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/lanthora/uranus/internal/web/process"
	"github.com/sirupsen/logrus"
)

type NotifyWorker struct {
	running            bool
	wg                 sync.WaitGroup
	username           string
	password           string
	server             string
	client             *http.Client
	ProcessEventOffset int64
}

const notifyNumberMax = 10

func NewWorker(server, username, password string, processEventOffset int64) *NotifyWorker {
	w := NotifyWorker{
		server:             server,
		username:           username,
		password:           password,
		ProcessEventOffset: processEventOffset,
	}
	return &w
}

func (w *NotifyWorker) Start() (err error) {
	w.running = true

	jar, err := cookiejar.New(nil)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	w.client = &http.Client{Jar: jar}

	body, err := json.Marshal(map[string]string{"username": w.username, "password": w.password})
	if err != nil {
		logrus.Fatal(err)
		return
	}

	url := w.server + "/auth/login"
	contentType := "application/json"
	resp, err := w.client.Post(url, contentType, bytes.NewBuffer(body))
	if err != nil {
		logrus.Fatal(err)
		return
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Infof("Login: %s", bytes)

	w.wg.Add(1)
	go w.run()
	return
}
func (w *NotifyWorker) notify(title, message string) {
	err := beeep.Notify(title, message, "")
	if err != nil {
		logrus.Error(err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func (w *NotifyWorker) run() {
	defer w.wg.Done()
	for w.running {
		body, err := json.Marshal(map[string]int64{"offset": w.ProcessEventOffset, "limit": 5000})
		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return
		}

		url := w.server + "/process/listEvents"
		contentType := "application/json"
		resp, err := w.client.Post(url, contentType, bytes.NewBuffer(body))
		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return
		}

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return
		}

		doc := struct {
			Status  int             `json:"status"`
			Message string          `json:"message"`
			Data    []process.Event `json:"data"`
		}{}
		err = json.Unmarshal(bytes, &doc)
		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return
		}
		if doc.Data == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		for idx, event := range doc.Data {
			w.ProcessEventOffset = event.ID
			title := fmt.Sprintf("进程防护事件 (ID: %d)", event.ID)
			message := event.Argv
			w.notify(title, message)

			if idx > notifyNumberMax {
				break
			}
		}

		if w.ProcessEventOffset != doc.Data[len(doc.Data)-1].ID {
			title := "进程防护事件"
			message := "已忽略积压的通知,请通过网页查看"
			w.notify(title, message)

			event := doc.Data[len(doc.Data)-1]
			w.ProcessEventOffset = event.ID
			title = fmt.Sprintf("进程防护事件 (ID: %d)", w.ProcessEventOffset)
			message = event.Argv
			w.notify(title, message)
			break
		}
	}
}

func (w *NotifyWorker) Stop() {
	w.running = false
	w.wg.Wait()

	url := w.server + "/auth/logout"
	contentType := "application/json"
	resp, err := w.client.Post(url, contentType, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Infof("Logout: %s", bytes)
}
