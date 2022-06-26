package file

import (
	"encoding/json"
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

type Policy struct {
	ID        uint64 `json:"id" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Fsid      uint64 `json:"fsid" binding:"required"`
	Ino       uint64 `json:"ino" binding:"required"`
	Perm      int    `json:"perm" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
	Status    int    `json:"status" binding:"required"`
}

type Event struct {
	ID        uint64 `json:"id" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Fsid      uint64 `json:"fsid" binding:"required"`
	Ino       uint64 `json:"ino" binding:"required"`
	Perm      int    `json:"perm" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
	Policy    uint64 `json:"policy" binding:"required"`
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

func FileEnable() bool {
	tmp, err := connector.Exec(`{"type":"user::file::enable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code" binding:"required"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(tmp), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func FileDisable() bool {
	tmp, err := connector.Exec(`{"type":"user::file::disable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code" binding:"required"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(tmp), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func FileClear() bool {
	tmp, err := connector.Exec(`{"type":"user::file::clear"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code" binding:"required"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(tmp), &response); err != nil {
		return false
	}
	return response.Code == 0
}
