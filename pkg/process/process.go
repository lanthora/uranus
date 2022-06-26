package process

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"uranus/pkg/connector"
)

const (
	StatusDisable = 0
	StatusEnable  = 1
)

const (
	StatusJudgeDisable = 0
	StatusJudgeAudit   = 1
	StatusJudgeProtect = 2
)

const (
	StatusPending   = 0
	StatusUntrusted = 1
	StatusTrusted   = 2
)

func SplitCmd(cmd string) (workdir string, binary string, argv string) {
	raw := strings.Split(cmd, "\u001f")
	workdir = raw[0]
	binary = raw[1]
	argv = raw[2]
	for i := 3; i < len(raw); i++ {
		argv += fmt.Sprintf(" %s", raw[i])
	}
	return
}

func ProcessJudgeUpdate(judge int) bool {
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
		Code  int         `json:"code" binding:"required"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func ProcessEnable() bool {
	responseStr, err := connector.Exec(`{"type":"user::proc::enable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code" binding:"required"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func ProcessDisable() bool {
	responseStr, err := connector.Exec(`{"type":"user::proc::disable"}`, time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code" binding:"required"`
		Extra interface{} `json:"extra"`
	}{}
	if err := json.Unmarshal([]byte(responseStr), &response); err != nil {
		return false
	}
	return response.Code == 0
}
