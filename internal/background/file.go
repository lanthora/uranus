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
	"uranus/pkg/file"
	"uranus/pkg/watchdog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

const (
	sqlCreateFilePolicyTable       = `create table if not exists file_policy(id integer primary key autoincrement, path text not null, fsid integer, ino integer, perm integer not null, timestamp integer not null, status integer not null)`
	sqlCreateFileEventTable        = `create table if not exists file_event(id integer primary key autoincrement, path text not null, fsid integer, ino integer, perm integer not null, timestamp integer not null, policy integer not null)`
	sqlQueryFilePolicy             = `select id,path,fsid,ino,perm from file_policy`
	sqlUpdateFilePolicyFsidInoById = `update file_policy set fsid=?,ino=?,timestamp=? where id=?`
	sqlUpdateFilePolicyStatusById  = `update file_policy set status=? where id=?`
	sqlQueryFilePolicyIdByFsidIno  = `select id from file_policy where fsid=? and ino=? and status=0`
	sqlInsertFileEvent             = `insert into file_event(path,fsid,ino,perm,timestamp,policy) values(?,?,?,?,?,?)`
)

type FileWorker struct {
	dataSourceName string

	running bool
	wg      sync.WaitGroup
	conn    connector.Connector
	config  *config.Config
	dog     *watchdog.Watchdog
}

func NewFileWorker(dataSourceName string) *FileWorker {
	worker := FileWorker{
		dataSourceName: dataSourceName,
	}
	return &worker
}

func (w *FileWorker) Start() (err error) {
	w.running = true
	err = w.conn.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}
	err = w.initDB()
	if err != nil {
		logrus.Error(err)
		return
	}

	err = w.conn.Send(`{"type":"user::msg::sub","section":"kernel::file::report"}`)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = w.conn.Send(`{"type":"user::msg::sub","section":"osinfo::report"}`)
	if err != nil {
		logrus.Error(err)
		return
	}

	if err = w.initFilePolicy(); err != nil {
		logrus.Error(err)
		return
	}

	coreStatus, tmpErr := w.config.GetInteger("file::core::status")
	if tmpErr == nil && coreStatus == file.StatusEnable {
		if err = w.conn.Send(`{"type":"user::file::enable"}`); err != nil {
			logrus.Error(err)
			return
		}
	}

	w.wg.Add(1)
	go w.run()
	return
}

func (w *FileWorker) Stop() (err error) {
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"kernel::proc::report"}`)
	if err != nil {
		return
	}
	err = w.conn.Send(`{"type":"user::msg::unsub","section":"osinfo::report"}`)
	if err != nil {
		return
	}

	err = w.conn.Send(`{"type":"user::file::disable"}`)
	if err != nil {
		return
	}

	err = w.conn.Send(`{"type":"user::file::clear"}`)
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

func (w *FileWorker) handleMsg(msg string) {
	event := struct {
		Type string `json:"type"`
		Path string `json:"name"`
		Fsid uint64 `json:"fsid"`
		Ino  uint64 `json:"ino"`
		Perm int    `json:"perm"`
	}{}

	err := json.Unmarshal([]byte(msg), &event)
	if err != nil {
		logrus.Error(err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}
	switch event.Type {
	case "kernel::file::report":
		err = w.handleFileEvent(event.Path, event.Fsid, event.Ino, event.Perm)
		if err != nil {
			logrus.Error(err)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	case "osinfo::report":
		w.dog.Kick()
	}
}

func (w *FileWorker) run() {
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

func (w *FileWorker) initDB() (err error) {
	os.MkdirAll(filepath.Dir(w.dataSourceName), os.ModeDir)
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec(sqlCreateFilePolicyTable)
	if err != nil {
		logrus.Error(err)
		return
	}

	_, err = db.Exec(sqlCreateFileEventTable)
	if err != nil {
		logrus.Error(err)
		return
	}

	w.config, err = config.New(w.dataSourceName)
	if err != nil {
		return
	}
	return
}

func (w *FileWorker) setPolicyThenGetExceptionPolicies() (policies []file.Policy, err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryFilePolicy)
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
		var (
			policy file.Policy
			fsid   uint64
			ino    uint64
			status int
		)
		err = rows.Scan(&policy.ID, &policy.Path, &policy.Fsid, &policy.Ino, &policy.Perm)
		if err != nil {
			logrus.Error(err)
			return
		}

		fsid, ino, status, err = file.SetPolicy(policy.Path, policy.Perm, file.FlagNew)
		if err != nil {
			logrus.Error(err)
			return
		}
		if policy.Fsid != fsid || policy.Ino != ino || policy.Status != status {
			policy.Fsid = fsid
			policy.Ino = ino
			policy.Status = status
			policies = append(policies, policy)
		}
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func (w *FileWorker) initFilePolicy() (err error) {
	policies, err := w.setPolicyThenGetExceptionPolicies()
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, policy := range policies {
		err = w.updateFilePolcyFsidInoById(policy.Fsid, policy.Ino, policy.ID)
		if err != nil {
			logrus.Error(err)
			return
		}
		err = w.updateFilePolcyStatusById(policy.Status, policy.ID)
		if err != nil {
			logrus.Error(err)
			return
		}
	}
	return
}

func (w *FileWorker) updateFilePolcyFsidInoById(fsid, ino, id uint64) (err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlUpdateFilePolicyFsidInoById)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(fsid, ino, time.Now().Unix(), id)
	if err != nil {
		logrus.Error(err)
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		logrus.Error(err)
		return
	}
	if affected != 1 {
		logrus.Errorf("id=%d, fsid=%d, ino=%d, affected=%d", id, fsid, ino, affected)
	}
	return
}

func (w *FileWorker) updateFilePolcyStatusById(status int, id uint64) (err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlUpdateFilePolicyStatusById)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(status, id)
	if err != nil {
		logrus.Error(err)
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		logrus.Error(err)
		return
	}
	if affected != 1 {
		logrus.Errorf("id=%d, status=%d, affected=%d", id, status, affected)
	}
	return
}

func (w *FileWorker) handleFileEvent(path string, fsid, ino uint64, perm int) (err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryFilePolicyIdByFsidIno)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()

	var policyId uint64
	err = stmt.QueryRow(fsid, ino).Scan(&policyId)
	if err != nil {
		logrus.Error(err)
		return
	}

	stmt, err = db.Prepare(sqlInsertFileEvent)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(path, fsid, ino, perm, time.Now().Unix(), policyId)
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}
