// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var (
	sqlCreateConfigTable = `create table if not exists config(id integer primary key autoincrement, key text not null unique, integer integer, real real, text text)`
	sqlInsertInteger     = `insert into config(key,integer) values(?,?)`
	sqlUpdateInteger     = `update config set integer=? where key=?`
	sqlQueryInteger      = `select integer from config where key=?`
	sqlInsertReal        = `insert into config(key,real) values(?,?)`
	sqlUpdateReal        = `update config set real=? where key=?`
	sqlQueryReal         = `select real from config where key=?`
	sqlInsertText        = `insert into config(key,text) values(?,?)`
	sqlUpdateText        = `update config set text=? where key=?`
	sqlQueryText         = `select text from config where key=?`
)

type Config struct {
	dbName string
}

func New(dbName string) (c *Config, err error) {
	c = &Config{
		dbName: dbName,
	}
	os.MkdirAll(filepath.Dir(c.dbName), os.ModeDir)
	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec(sqlCreateConfigTable)
	if err != nil {
		return
	}
	return
}

func (c *Config) SetInteger(key string, value int) (err error) {

	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlUpdateInteger)
	if err != nil {
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(value, key)
	if err != nil {
		return
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		return
	}

	stmt, err = db.Prepare(sqlInsertInteger)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(key, value)
	if err != nil {
		return
	}

	return
}

func (c *Config) GetInteger(key string) (value int, err error) {
	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryInteger)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(key).Scan(&value)
	return
}

func (c *Config) SetReal(key string, value float64) (err error) {

	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlUpdateReal)
	if err != nil {
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(value, key)
	if err != nil {
		return
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		return
	}

	stmt, err = db.Prepare(sqlInsertReal)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(key, value)
	if err != nil {
		return
	}

	return
}

func (c *Config) GetReal(key string) (value float64, err error) {
	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryReal)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(key).Scan(&value)
	return
}

func (c *Config) SetText(key string, value string) (err error) {

	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlUpdateText)
	if err != nil {
		return
	}
	defer stmt.Close()
	result, err := stmt.Exec(value, key)
	if err != nil {
		return
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		return
	}

	stmt, err = db.Prepare(sqlInsertText)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(key, value)
	if err != nil {
		return
	}

	return
}

func (c *Config) GetText(key string) (value string, err error) {
	db, err := sql.Open("sqlite3", c.dbName)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryText)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(key).Scan(&value)
	return
}
