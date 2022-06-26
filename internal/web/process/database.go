// SPDX-License-Identifier: AGPL-3.0-or-later
package process

import (
	"database/sql"
)

const (
	sqlQueryProcessLimitOffset = `select id,workdir,binary,argv,count,judge,status from process_event limit ? offset ?`
	sqlUpdateProcessStatus     = `update process_event set status=? where id=?`
	sqlQueryProcessCmdById     = `select cmd from process_event where id=?`
)

func (w *Worker) queryLimitOffset(limit, offset int) (events []Event, err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryProcessLimitOffset)
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
		e := Event{}
		err = rows.Scan(&e.ID, &e.Workdir, &e.Binary, &e.Argv, &e.Count, &e.Judge, &e.Status)
		if err != nil {
			return
		}
		events = append(events, e)
	}
	err = rows.Err()
	return
}

func (w *Worker) updateStatus(id uint64, status int) bool {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return false
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlUpdateProcessStatus)
	if err != nil {
		return false
	}
	defer stmt.Close()

	result, err := stmt.Exec(status, id)
	if err != nil {
		return false
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false
	}
	return affected == 1
}

func (w *Worker) queryCmdById(id int) (cmd string, err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryProcessCmdById)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&cmd)
	return
}
