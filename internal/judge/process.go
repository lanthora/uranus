// SPDX-License-Identifier: AGPL-3.0-or-later
package judge

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"uranus/pkg/connector"
	"uranus/pkg/watchdog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

const (
	sqlCreateProcessJudgeTable = `create table if not exists judge(id integer primary key autoincrement, cmd blob not null, times integer default 1)`
	sqlQueryTimesByCmd         = `select id, times from judge where cmd=? limit 1`
	sqlQueryCmdByTimes         = `select cmd from judge where times >= ?`
	sqlInsertCmd               = `insert into judge(cmd) values(?)`
	sqlIncreCmdTimes           = `update judge set times=times+1 where cmd=?`
)

type ProcessWorker struct {
	dbName  string
	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
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

	_, err = db.Exec(sqlCreateProcessJudgeTable)
	if err != nil {
		return
	}
	return
}

func (w *ProcessWorker) getCmdTimes(cmd string) (times int, err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryTimesByCmd)
	if err != nil {
		return
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(cmd).Scan(&id, &times)
	if err == sql.ErrNoRows {
		times = 0
		err = nil
		return
	}
	return
}

func (w *ProcessWorker) increCmdTimes(cmd string) (err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlIncreCmdTimes)
	if err != nil {
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(cmd)
	if err != nil {
		return
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		return
	}

	stmt, err = db.Prepare(sqlInsertCmd)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(cmd)
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

	stmt, err := db.Prepare(sqlQueryCmdByTimes)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(3)
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

func (w *ProcessWorker) run() {
	defer w.wg.Done()
	dog := watchdog.New(10*time.Second, func() {
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

		var doc map[string]interface{}
		err = json.Unmarshal([]byte(msg), &doc)
		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			continue
		}
		switch doc["type"].(string) {
		case "kernel::proc::report":
			cmd := doc["cmd"].(string)
			err := w.increCmdTimes(cmd)
			if err != nil {
				logrus.Error(err)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		case "audit::proc::report":
			cmd := doc["cmd"].(string)
			times, err := w.getCmdTimes(cmd)
			if err != nil {
				logrus.Error(err)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				break
			}
			if times < 3 {
				break
			}
			err = w.setTrustedCmd(cmd)
			if err != nil {
				logrus.Error(err)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		case "osinfo::report":
			dog.Kick()
		}
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

	err = w.conn.Send(`{"type":"user::proc::enable"}`)
	if err != nil {
		return
	}

	err = w.conn.Send(`{"type":"user::msg::sub","section":"audit::proc::report"}`)
	if err != nil {
		return
	}
	err = w.conn.Send(`{"type":"user::msg::sub","section":"kernel::proc::report"}`)
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
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"kernel::proc::report"}`)
	if err != nil {
		return
	}
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
