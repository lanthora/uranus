package file

import (
	"database/sql"
	"time"
	"uranus/pkg/file"

	_ "github.com/mattn/go-sqlite3"
)

const (
	sqlInsertFilePolicy           = `insert into file_policy(path,fsid,ino,perm,timestamp,status) values(?,?,?,?,?,?)`
	sqlUpdateFilePolicyById       = `update file_policy set fsid=?,ino=?,perm=?,timestamp=?,status=? where id=?`
	sqlQueryFileEventLimitOffset  = `select id,path,fsid,ino,perm,timestamp,policy from file_event limit ? offset ?`
	sqlQueryFilePolicyById        = `select id,path,fsid,ino,perm,timestamp,status from file_policy where id=?`
	sqlQueryFilePolicyLimitOffset = `select id,path,fsid,ino,perm,timestamp,status from file_policy limit ? offset ?`
	sqlDeleteFilePolicyById       = `delete from file_policy where id=?`
	sqlDeleteFileEventById        = `delete from file_event where id=?`
)

func (w *Worker) insertFilePolicy(path string, fsid, ino uint64, perm, status int) (err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlInsertFilePolicy)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(path, fsid, ino, perm, time.Now().Unix(), status)
	if err != nil {
		return
	}
	return
}

func (w *Worker) updateFilePolicyById(fsid, ino uint64, perm, status, id int) (err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlUpdateFilePolicyById)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(fsid, ino, perm, time.Now().Unix(), status, id)
	if err != nil {
		return
	}
	return
}

func (w *Worker) queryFileEventOffsetLimit(limit, offset int) (events []file.Event, err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryFileEventLimitOffset)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		e := file.Event{}
		err = rows.Scan(&e.ID, &e.Path, &e.Fsid, &e.Ino, &e.Perm, &e.Timestamp, &e.Policy)
		if err != nil {
			return
		}
		events = append(events, e)
	}
	err = rows.Err()
	return
}

func (w *Worker) queryFilePolicyById(id int) (event file.Policy, err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryFilePolicyById)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&event.ID, &event.Path, &event.Fsid, &event.Ino, &event.Perm, &event.Timestamp, &event.Status)
	return
}

func (w *Worker) queryFilePolicyLimitOffset(limit, offset int) (events []file.Policy, err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryFilePolicyLimitOffset)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		e := file.Policy{}
		err = rows.Scan(&e.ID, &e.Path, &e.Fsid, &e.Ino, &e.Perm, &e.Timestamp, &e.Status)
		if err != nil {
			return
		}
		events = append(events, e)
	}
	err = rows.Err()
	return
}

func (w *Worker) deleteFilePolicyById(id int) (err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlDeleteFilePolicyById)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return
	}
	return
}

func (w *Worker) deleteFileEventById(id int) (err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlDeleteFileEventById)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return
	}
	return
}
