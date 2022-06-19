package process

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"uranus/pkg/connector"
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
	// 参数组装成底层命令
	request := map[string]interface{}{
		"type":  "user::proc::judge",
		"judge": judge,
	}

	bytes, err := json.Marshal(request)
	if err != nil || len(bytes) > 1024 {
		return false
	}

	// 转换成字符串向底层发送命令,并接收响应的字符串
	responseStr, err := connector.Exec(string(bytes), time.Second)
	if err != nil {
		return false
	}

	// 底层返回的字符串转换后返回给前端
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
