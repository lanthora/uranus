// SPDX-License-Identifier: AGPL-3.0-or-later
package process

import (
	"database/sql"
)

const (
	// 提供给前端的可分页的查询功能
	sqlQueryLimitOffset = `select id,workdir,binary,argv,count,judge,status from process limit ? offset ?`
	// 前端发送请求更新处理方式
	sqlUpdateStatus = `update process set status=? where id=?`
	// 如果处理方式是设置为信任,需要从数据库查出 cmd 并设置给底层
	sqlQueryCmdById = `select cmd from process where id=?`
)

func (w *Worker) queryLimitOffset(limit, offset int) (events []Event, err error) {
	db, err := sql.Open("sqlite3", w.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryLimitOffset)
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

	stmt, err := db.Prepare(sqlUpdateStatus)
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
	stmt, err := db.Prepare(sqlQueryCmdById)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&cmd)
	return
}
