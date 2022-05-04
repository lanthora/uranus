package judge

import (
	"database/sql"
	"encoding/json"
	"sync"
	"syscall"
	"time"
	"uranus/pkg/connector"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

const (
	sqlCreateProcessJudgeTable = `create table if not exists judge(id integer primary key autoincrement, cmd blob not null, times integer default 1)`
	sqlQueryTimesByCmd         = `select id, times from judge where cmd=? limit 1`
	sqlQueryCmdByTimes         = `select cmd from judge where times >= ?`
	sqlInsertCmd               = `insert into judge(cmd) values(?)`
	sqlUpdateCmdTimes          = `update judge set times=? where id=?`
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
	var times int
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
	if err != nil && err != sql.ErrNoRows {
		return
	}

	if err == sql.ErrNoRows {
		times = 1
		stmt, err = db.Prepare(sqlInsertCmd)
		if err != nil {
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(cmd)
		if err != nil {
			return
		}
	} else {
		times = times + 1
		stmt, err = db.Prepare(sqlUpdateCmdTimes)
		if err != nil {
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(times, id)
		if err != nil {
			return
		}
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

		if doc["type"].(string) == "kernel::proc::report" {
			cmd := doc["cmd"].(string)
			err := w.increCmdTimes(cmd)
			if err != nil {
				logrus.Error(err)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				continue
			}
		} else if doc["type"].(string) == "audit::proc::report" {
			cmd := doc["cmd"].(string)
			times, err := w.getCmdTimes(cmd)
			if err != nil {
				logrus.Error(err)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				continue
			}
			if times < 3 {
				continue
			}
			err = w.setTrustedCmd(cmd)
			if err != nil {
				logrus.Error(err)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		}
	}
	return
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
	err = w.conn.Send(`{"type":"user::proc::disable"}`)
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	w.running = false
	err = w.conn.Shutdown()
	if err != nil {
		return
	}
	w.wg.Wait()
	w.conn.Close()
	return
}
