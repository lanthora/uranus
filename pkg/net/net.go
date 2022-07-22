// SPDX-License-Identifier: AGPL-3.0-or-later
package net

import (
	"encoding/json"
	"errors"
	"time"
	"uranus/pkg/connector"
)

const (
	StatusDisable = 0
	StatusEnable  = 1
)

var (
	ErrorEnable      = errors.New("net protection enable failed")
	ErrorDisable     = errors.New("net protection disable failed")
	ErrorClearPolicy = errors.New("clear net policy failed")
)

type Policy struct {
	ID       int64 `json:"id"`
	Priority int8  `json:"priority"`
	Addr     struct {
		Src struct {
			Begin string `json:"begin"`
			End   string `json:"end"`
		} `json:"src"`
		Dst struct {
			Begin string `json:"begin"`
			End   string `json:"end"`
		} `json:"dst"`
	} `json:"addr"`
	Protocol struct {
		Begin uint8 `json:"begin"`
		End   uint8 `json:"end"`
	} `json:"protocol"`
	Port struct {
		Src struct {
			Begin uint16 `json:"begin"`
			End   uint16 `json:"end"`
		} `json:"src"`
		Dst struct {
			Begin uint16 `json:"begin"`
			End   uint16 `json:"end"`
		} `json:"dst"`
	} `json:"port"`
	Flags    int32  `json:"flags"`
	Response uint32 `json:"response"`
}

func AddPolicy(policy Policy) bool {
	request := struct {
		Type string `json:"type"`
		*Policy
	}{
		Type:   "user::net::insert",
		Policy: &policy,
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		return false
	}

	tmp, err := connector.Exec(string(bytes), time.Second)
	if err != nil {
		return false
	}

	response := struct {
		Code  int         `json:"code"`
		Extra interface{} `json:"extra"`
	}{}
	if err = json.Unmarshal([]byte(tmp), &response); err != nil {
		return false
	}
	return response.Code == 0
}

func DeletePolicy(id int) bool {
	request := struct {
		Type string `json:"type"`
		ID   int    `json:"id"`
	}{
		Type: "user::net::delete",
		ID:   id,
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		return false
	}

	tmp, err := connector.Exec(string(bytes), time.Second)
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

func Enable() bool {
	tmp, err := connector.Exec(`{"type":"user::net::enable"}`, time.Second)
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
	tmp, err := connector.Exec(`{"type":"user::net::disable"}`, time.Second)
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
	tmp, err := connector.Exec(`{"type":"user::net::clear"}`, time.Second)
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
