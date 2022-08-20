// SPDX-License-Identifier: AGPL-3.0-or-later
package background

import (
	"database/sql"
	"encoding/json"
	"sync"
	"syscall"
	"time"
	"uranus/internal/config"
	"uranus/pkg/connector"
	"uranus/pkg/process"
	"uranus/pkg/watchdog"

	"github.com/sirupsen/logrus"
)

const (
	sqlCreateProcessTable    = `create table if not exists process_event(id integer primary key autoincrement, cmd blob not null unique, workdir text not null, binary text not null, argv text not null, count integer not null, judge integer not null, status integer not null)`
	sqlCreateProcessCmdIndex = `create unique index if not exists process_cmd_idx on process_event (cmd)`
	sqlUpdateProcessCount    = `update process_event set count=count+1,judge=?,status=? where cmd=?`
	sqlInsertProcessEvent    = `insert into process_event(cmd,workdir,binary,argv,count,judge,status) values(?,?,?,?,1,?,?)`
	sqlQueryAllowedProcesses = `select cmd from process_event where status=2`
)

type ProcessWorker struct {
	db *sql.DB

	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
	config  *config.Config
	dog     *watchdog.Watchdog
}

func NewProcessWorker(db *sql.DB) *ProcessWorker {
	worker := ProcessWorker{
		db: db,
	}
	return &worker
}

func (w *ProcessWorker) Init() (err error) {
	err = w.initDB()
	if err != nil {
		return
	}

	w.config, err = config.New(w.db)
	if err != nil {
		logrus.Error(err)
		return
	}

	err = w.initTrustedCmd()
	if err != nil {
		return
	}

	status, err := w.config.GetInteger(config.ProcessModuleStatus)
	if err != nil {
		err = nil
		status = process.StatusDisable
	}

	switch status {
	case process.StatusEnable:
		if ok := process.Enable(); !ok {
			err = process.ErrorEnable
			return
		}
	default:
		if ok := process.Disable(); !ok {
			err = process.ErrorEnable
			return
		}
	}

	judge, err := w.config.GetInteger(config.ProcessProtectionMode)
	if err != nil {
		err = nil
		judge = process.StatusJudgeDisable
	}

	if ok := process.UpdateJudge(judge); !ok {
		err = process.ErrorUpdateJudge
		return
	}

	return
}

func (w *ProcessWorker) Start() (err error) {
	w.running = true
	err = w.conn.Connect()
	if err != nil {
		return
	}

	err = w.conn.Send(`{"type":"user::msg::sub","section":"audit::proc::report"}`)
	if err != nil {
		return
	}
	err = w.conn.Send(`{"type":"user::msg::sub","section":"osinfo::report"}`)
	if err != nil {
		return
	}

	w.wg.Add(1)
	go w.run()
	return
}

func (w *ProcessWorker) Stop() {
	err := w.conn.Send(`{"type":"user::msg::unsub","section":"audit::proc::report"}`)
	if err != nil {
		logrus.Error(err)
	}
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"osinfo::report"}`)
	if err != nil {
		logrus.Error(err)
	}

	if ok := process.Disable(); !ok {
		logrus.Error("process protection disable failed")
	}

	if ok := process.ClearPolicy(); !ok {
		logrus.Error("process protection clear failed")
	}

	time.Sleep(time.Second)
	w.running = false
	err = w.conn.Shutdown(time.Now())
	if err != nil {
		logrus.Error(err)
	}
	w.wg.Wait()
	w.conn.Close()
}

func (w *ProcessWorker) initDB() (err error) {
	_, err = w.db.Exec(sqlCreateProcessTable)
	if err != nil {
		logrus.Error(err)
		return
	}

	_, err = w.db.Exec(sqlCreateProcessCmdIndex)
	if err != nil {
		logrus.Error(err)
		return
	}

	return
}

func (w *ProcessWorker) initTrustedCmd() (err error) {
	stmt, err := w.db.Prepare(sqlQueryAllowedProcesses)
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
		cmd := ""
		err = rows.Scan(&cmd)
		if err != nil {
			return
		}
		process.SetTrustedCmd(cmd)
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func (w *ProcessWorker) updateCmd(cmd string, judge int) (err error) {
	status, err := w.config.GetInteger(config.ProcessCmdDefaultStatus)
	if err != nil {
		status = process.StatusPending
	}
	if status == process.StatusTrusted && judge != process.StatusJudgeDefense {
		process.SetTrustedCmd(cmd)
	}

	stmt, err := w.db.Prepare(sqlUpdateProcessCount)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(judge, status, cmd)
	if err != nil {
		logrus.Error(err)
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		logrus.Error(err)
		return
	}

	if affected != 0 {
		return
	}

	stmt, err = w.db.Prepare(sqlInsertProcessEvent)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()
	workdir, binary, argv, err := process.SplitCmd(cmd)
	if err != nil {
		logrus.Error(err)
		return
	}
	_, err = stmt.Exec(cmd, workdir, binary, argv, judge, status)
	if err != nil {
		logrus.Error(err)
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
		}
	default:
	}
}

func (w *ProcessWorker) run() {
	defer w.wg.Done()
	w.dog = watchdog.New(10*time.Second, func() {
		logrus.Error("osinfo::report timeout")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	})
	defer w.dog.Stop()
	for w.running {
		msg, err := w.conn.Recv()

		if !w.running {
			logrus.Info("process worker exit")
			break
		}

		if err != nil {
			logrus.Error(err)
			continue
		}
		w.dog.Kick()
		go w.handleMsg(msg)
	}
}
