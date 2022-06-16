package user

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var (
	userDB string
)

var (
	sqlCreateUserTable         = `create table if not exists user(id integer primary key autoincrement, username text not null unique, salt text not null, password text not null, alias text, permissions text)`
	sqlInsertUser              = `insert into user(username, salt, password, alias, permissions) values(?,?,?,?,?)`
	sqlQueryUserCount          = `select count(*) from user`
	sqlQueryAllUser            = `select id, username, alias, permissions from user`
	sqlQueryUserByUsername     = `select id, alias, permissions from user where username=?`
	sqlQueryPasswordByUsername = `select salt, password from user where username=?`
	sqlUpdateUser              = `update user set username=?, salt=?, password=?, alias=?, permissions=? where id=?`
	sqlDeleteUser              = `delete from user where id=?`
)

type UserInDb struct {
}

func initUserTable(dbName string) (err error) {
	os.MkdirAll(filepath.Dir(dbName), os.ModeDir)
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec(sqlCreateUserTable)
	if err != nil {
		return
	}
	userDB = dbName
	return
}

func noUser() (ok bool, err error) {
	db, err := sql.Open("sqlite3", userDB)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryUserCount)
	if err != nil {
		return
	}
	defer stmt.Close()

	var count int
	if err = stmt.QueryRow().Scan(&count); err != nil {
		return
	}
	ok = (count == 0)
	return
}

func createUser(username, password, alias, permissions string) (err error) {
	salt := uuid.NewString()
	sum := sha256.Sum256([]byte(salt + password))
	hash := hex.EncodeToString(sum[:])

	db, err := sql.Open("sqlite3", userDB)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(sqlInsertUser)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, salt, hash, alias, permissions)
	if err != nil {
		return
	}
	return

}

func checkUserPassword(username, password string) (ok bool, err error) {
	db, err := sql.Open("sqlite3", userDB)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryPasswordByUsername)
	if err != nil {
		return
	}
	defer stmt.Close()

	var salt, hash string
	if err = stmt.QueryRow(username).Scan(&salt, &hash); err != nil {
		return
	}
	sum := sha256.Sum256([]byte(salt + password))
	ok = (hash == hex.EncodeToString(sum[:]))
	return
}

func queryUserByUsername(username string) (user User, err error) {
	user.Username = username
	db, err := sql.Open("sqlite3", userDB)
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlQueryUserByUsername)
	if err != nil {
		return
	}
	defer stmt.Close()

	if err = stmt.QueryRow(user.Username).Scan(&user.UserID, &user.AliasName, &user.Permissions); err != nil {
		return
	}
	return
}
