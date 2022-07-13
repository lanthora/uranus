// SPDX-License-Identifier: AGPL-3.0-or-later
package background

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"uranus/internal/config"
	"uranus/pkg/connector"
	"uranus/pkg/net"
	"uranus/pkg/watchdog"

	"github.com/sirupsen/logrus"
)

const (
	sqlCreateNetPolicyTable = `create table if not exists net_policy(id integer primary key autoincrement, priority int8, addr_src_begin text, addr_src_end text, addr_dst_begin text, addr_dst_end text, protocol_begin int, protocol_end int, port_src_begin int, port_src_end int, port_dst_begin int, port_dst_end int, flags int, response int)`
	sqlQueryNetPolicy       = `select id,priority,addr_src_begin,addr_src_end,addr_dst_begin,addr_dst_end,protocol_begin,protocol_end,port_src_begin,port_src_end,port_dst_begin,port_dst_end,flags,response from net_policy`
)

type NetWorker struct {
	dataSourceName string

	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
	config  *config.Config
	dog     *watchdog.Watchdog
}

func NewNetWorker(dataSourceName string) *NetWorker {
	worker := NetWorker{
		dataSourceName: dataSourceName,
	}
	return &worker
}

func (w *NetWorker) Init() (err error) {
	err = w.initDB()
	if err != nil {
		return
	}

	w.config, err = config.New(w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}

	if err = w.initNetPolicy(); err != nil {
		logrus.Error(err)
		return
	}

	return
}
func (w *NetWorker) Start() (err error) {

	w.running = true
	err = w.conn.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	err = w.conn.Send(`{"type":"user::msg::sub","section":"osinfo::report"}`)
	if err != nil {
		return
	}

	status, err := w.config.GetInteger(config.NetModuleStatus)
	if err != nil {
		err = nil
		return
	}

	if status != net.StatusEnable {
		return
	}

	if ok := net.Enable(); !ok {
		logrus.Error(net.EnableError)
	}

	w.wg.Add(1)
	go w.run()
	return
}

func (w *NetWorker) Stop() (err error) {
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"osinfo::report"}`)
	if err != nil {
		return
	}

	if ok := net.ClearPolicy(); !ok {
		logrus.Error(net.ClearPolicyError)
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

func (w *NetWorker) initDB() (err error) {
	os.MkdirAll(filepath.Dir(w.dataSourceName), os.ModeDir)
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec(sqlCreateNetPolicyTable)
	if err != nil {
		logrus.Error(err)
		return
	}

	return
}

func (w *NetWorker) initNetPolicy() (err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryNetPolicy)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		logrus.Error(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		policy := net.Policy{}
		err = rows.Scan(&policy.ID, &policy.Priority,
			&policy.Addr.Src.Begin, &policy.Addr.Src.End,
			&policy.Addr.Dst.Begin, &policy.Addr.Dst.End,
			&policy.Protocol.Begin, &policy.Protocol.End,
			&policy.Port.Src.Begin, &policy.Port.Src.End,
			&policy.Port.Dst.Begin, &policy.Port.Dst.End,
			&policy.Flags, &policy.Response)
		if err != nil {
			logrus.Error(err)
			return
		}
		if ok := net.AddPolicy(policy); !ok {
			logrus.Error(err)
			return
		}
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func (w *NetWorker) handleMsg(msg string) {
	event := struct {
		Type string `json:"type"`
	}{}

	err := json.Unmarshal([]byte(msg), &event)
	if err != nil {
		logrus.Error(err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}
	switch event.Type {
	case "osinfo::report":
		w.dog.Kick()
	}
}

func (w *NetWorker) run() {
	defer w.wg.Done()
	w.dog = watchdog.New(10*time.Second, func() {
		logrus.Error("osinfo::report timeout")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	})
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
		go w.handleMsg(msg)

	}
}
