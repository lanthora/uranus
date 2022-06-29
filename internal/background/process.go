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
	"uranus/pkg/process"
	"uranus/pkg/watchdog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

const (
	sqlCreateProcessTable    = `create table if not exists process_event(id integer primary key autoincrement, cmd blob not null unique, workdir text not null, binary text not null, argv text not null, count integer not null, judge integer not null, status integer not null)`
	sqlCreateProcessCmdIndex = `create unique index if not exists process_cmd_idx on process_event (cmd)`
	sqlUpdateProcessCount    = `update process_event set count=count+1,judge=? where cmd=?`
	sqlInsertProcessEvent    = `insert into process_event(cmd,workdir,binary,argv,count,judge,status) values(?,?,?,?,1,?,0)`
	sqlQueryAllowedProcesses = `select cmd from process_event where status=2`
)

type ProcessWorker struct {
	dbName string

	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
	config  *config.Config
	dog     *watchdog.Watchdog
}

func NewProcessWorker(dbName string) *ProcessWorker {
	worker := ProcessWorker{
		dbName: dbName,
	}
	return &worker
}

func (w *ProcessWorker) initDB() (err error) {
	os.MkdirAll(filepath.Dir(w.dbName), os.ModeDir)
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec(sqlCreateProcessTable)
	if err != nil {
		return
	}

	_, err = db.Exec(sqlCreateProcessCmdIndex)
	if err != nil {
		return
	}

	w.config, err = config.New(w.dbName)
	if err != nil {
		return
	}
	return
}

func (w *ProcessWorker) setTrustedCmd(cmd string) (err error) {
	data := map[string]string{
		"type": "user::proc::trusted::insert",
		"cmd":  cmd,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	err = w.conn.Send(string(b))
	if err != nil {
		return
	}
	return
}

func (w *ProcessWorker) initTrustedCmd() (err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryAllowedProcesses)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cmd string
		err = rows.Scan(&cmd)
		if err != nil {
			return
		}
		w.setTrustedCmd(cmd)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func (w *ProcessWorker) updateCmd(cmd string, judge int) (err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlUpdateProcessCount)
	if err != nil {
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(judge, cmd)
	if err != nil {
		return
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		return
	}

	stmt, err = db.Prepare(sqlInsertProcessEvent)
	if err != nil {
		return
	}
	defer stmt.Close()
	workdir, binary, argv := process.SplitCmd(cmd)
	_, err = stmt.Exec(cmd, workdir, binary, argv, judge)
	if err != nil {
		return
	}
	return
}

func (w *ProcessWorker) handleMsg(msg string) {
	event := struct {
		Type  string `json:"type"`
		Cmd   string `json:"cmd"`
		Judge int    `json:"judge"`
	}{}

	err := json.Unmarshal([]byte(msg), &event)
	if err != nil {
		logrus.Error(err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}
	switch event.Type {
	case "audit::proc::report":
		err = w.updateCmd(event.Cmd, event.Judge)
		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	case "osinfo::report":
		w.dog.Kick()
	}
}

func (w *ProcessWorker) run() {
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

func (w *ProcessWorker) Start() (err error) {
	w.running = true
	err = w.conn.Connect()
	if err != nil {
		return
	}
	err = w.initDB()
	if err != nil {
		return
	}

	status, err := w.config.GetInteger("proc::core::status")
	if err == nil && status == process.StatusEnable {
		err = w.conn.Send(`{"type":"user::proc::enable"}`)
		if err != nil {
			return
		}
	}

	err = w.conn.Send(`{"type":"user::msg::sub","section":"audit::proc::report"}`)
	if err != nil {
		return
	}
	err = w.conn.Send(`{"type":"user::msg::sub","section":"osinfo::report"}`)
	if err != nil {
		return
	}

	err = w.initTrustedCmd()
	if err != nil {
		return
	}

	w.wg.Add(1)
	go w.run()
	return
}

func (w *ProcessWorker) Stop() (err error) {
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"audit::proc::report"}`)
	if err != nil {
		return
	}
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"osinfo::report"}`)
	if err != nil {
		return
	}

	err = w.conn.Send(`{"type":"user::proc::disable"}`)
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
