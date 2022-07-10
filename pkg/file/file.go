// SPDX-License-Identifier: AGPL-3.0-or-later
package file

import (
	"encoding/json"
	"errors"
	"time"
	"uranus/pkg/connector"

	"github.com/sirupsen/logrus"
)

const (
	StatusDisable = 0
	StatusEnable  = 1
)

const (
	FlagAny    = 0
	FlagNew    = 1
	FlagUpdate = 2
)

const (
	StatusNormal       = 0
	StatusUnknown      = 1
	StatusConflict     = 2
	StatusFileNotExist = 3
)

var (
	EnableError = errors.New("file protection enable failed")
)

type Policy struct {
	ID        uint64 `json:"id"`
	Path      string `json:"path"`
	Fsid      uint64 `json:"fsid"`
	Ino       uint64 `json:"ino"`
	Perm      int    `json:"perm"`
	Timestamp int64  `json:"timestamp"`
	Status    int    `json:"status"`
}

type Event struct {
	ID        uint64 `json:"id"`
	Path      string `json:"path"`
	Fsid      uint64 `json:"fsid"`
	Ino       uint64 `json:"ino"`
	Perm      int    `json:"perm"`
	Timestamp int64  `json:"timestamp"`
	Policy    uint64 `json:"policy"`
}

func SetPolicy(path string, perm, flag int) (fsid, ino uint64, status int, err error) {
	request := map[string]interface{}{
		"type": "user::file::set",
		"path": path,
		"perm": perm,
		"flag": flag,
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		return
	}

	tmp, err := connector.Exec(string(bytes), time.Second)
	if err != nil {
		return
	}

	response := struct {
		Code int    `json:"code"`
		Fsid uint64 `json:"fsid"`
		Ino  uint64 `json:"ino"`
	}{}
	if err = json.Unmarshal([]byte(tmp), &response); err != nil {
		logrus.Error(err)
		return
	}

	fsid = response.Fsid
	ino = response.Ino

	switch response.Code {
	case 0:
		status = StatusNormal
	case -2:
		status = StatusFileNotExist
	case -17:
		status = StatusConflict
	default:
		status = StatusUnknown
	}
	return
}

func Enable() bool {
	tmp, err := connector.Exec(`{"type":"user::file::enable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(tmp), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func Disable() bool {
	tmp, err := connector.Exec(`{"type":"user::file::disable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(tmp), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func ClearPolicy() bool {
	tmp, err := connector.Exec(`{"type":"user::file::clear"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(tmp), &response); err != nil {
		return false
	}
	return response.Code == 0
}
