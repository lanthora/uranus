// SPDX-License-Identifier: AGPL-3.0-or-later
package net

import (
	"database/sql"
	"uranus/pkg/net"

	"github.com/sirupsen/logrus"
)

const (
	sqlInsertNetPolicy           = `insert into net_policy(priority,addr_src_begin,addr_src_end,addr_dst_begin,addr_dst_end,protocol_begin,protocol_end,port_src_begin,port_src_end,port_dst_begin,port_dst_end,flags,response) values(?,?,?,?,?,?,?,?,?,?,?,?,?)`
	sqlDeleteNetPolicyById       = `delete from net_policy where id=?`
	sqlQueryNetPolicyLimitOffset = `select id,priority,addr_src_begin,addr_src_end,addr_dst_begin,addr_dst_end,protocol_begin,protocol_end,port_src_begin,port_src_end,port_dst_begin,port_dst_end,flags,response from net_policy limit ? offset ?`
)

func (w *Worker) insertNetPolicy(policy *net.Policy) (id int64, err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlInsertNetPolicy)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(policy.Priority,
		policy.Addr.Src.Begin, policy.Addr.Src.End,
		policy.Addr.Dst.Begin, policy.Addr.Dst.End,
		policy.Protocol.Begin, policy.Protocol.End,
		policy.Port.Src.Begin, policy.Port.Src.End,
		policy.Port.Dst.Begin, policy.Port.Dst.End,
		policy.Flags, policy.Response)
	if err != nil {
		logrus.Error(err)
		return
	}
	id, err = result.LastInsertId()
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func (w *Worker) deleteNetPolicyById(id int) (err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlDeleteNetPolicyById)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func (w *Worker) queryNetPolicyLimitOffset(limit, offset int) (policies []net.Policy, err error) {
	db, err := sql.Open("sqlite3", w.dataSourceName)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlQueryNetPolicyLimitOffset)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(limit, offset)
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
		policies = append(policies, policy)
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
	}
	return
}
