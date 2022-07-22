// SPDX-License-Identifier: AGPL-3.0-or-later
package process

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"uranus/pkg/connector"

	"github.com/sirupsen/logrus"
)

const (
	StatusDisable = 0
	StatusEnable  = 1
)

const (
	StatusJudgeDisable = 0
	StatusJudgeAudit   = 1
	StatusJudgeDefense = 2
)

const (
	StatusPending   = 0
	StatusUntrusted = 1
	StatusTrusted   = 2
)

var (
	ErrorEnable      = errors.New("process protection enable failed")
	ErrorDisable     = errors.New("process protection disable failed")
	ErrorUpdateJudge = errors.New("process protection update Judge failed")
	ErrorInvalidCmd  = errors.New("invalid cmd")
)

func SplitCmd(cmd string) (workdir, binary, argv string, err error) {
	raw := strings.Split(cmd, "\u001f")
	if len(raw) < 3 {
		err = ErrorInvalidCmd
		return
	}
	workdir = raw[0]
	binary = raw[1]
	argv = raw[2]
	for i := 3; i < len(raw); i++ {
		argv += fmt.Sprintf(" %s", raw[i])
	}
	return
}

func UpdateJudge(judge int) bool {
	request := map[string]interface{}{
		"type":  "user::proc::judge",
		"judge": judge,
	}

	bytes, err := json.Marshal(request)
	if err != nil || len(bytes) > 1024 {
		return false
	}

	responseStr, err := connector.Exec(string(bytes), time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func Enable() bool {
	responseStr, err := connector.Exec(`{"type":"user::proc::enable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func Disable() bool {
	responseStr, err := connector.Exec(`{"type":"user::proc::disable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func ClearPolicy() bool {
	responseStr, err := connector.Exec(`{"type":"user::proc::trusted::clear"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func SetTrustedCmd(cmd string) (err error) {
	data := map[string]string{
		"type": "user::proc::trusted::insert",
		"cmd":  cmd,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}

	// TODO: 更细致的判断是否执行成功
	_, err = connector.Exec(string(b), time.Second)
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func SetUntrustedCmd(cmd string) (err error) {
	data := map[string]string{
		"type": "user::proc::trusted::delete",
		"cmd":  cmd,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}

	// TODO: 更细致的判断是否执行成功
	_, err = connector.Exec(string(b), time.Second)
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}
