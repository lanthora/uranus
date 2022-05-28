// SPDX-License-Identifier: AGPL-3.0-or-later
package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
)

var (
	loggedUser *lru.Cache
)

type User struct {
	UserID      uint64   `json:"userID" binding:"required"`
	Username    string   `json:"username" binding:"required"`
	AliasName   string   `json:"aliasName" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

func Init(engine *gin.Engine) (err error) {
	engine.POST("/user/login", userLogin)
	engine.POST("/user/logout", userLogout)
	engine.POST("/user/insert", userInsert)
	engine.POST("/user/delete", userDelete)
	engine.POST("/user/update", userUpdate)
	engine.POST("/user/query", userQuery)

	loggedUser, err = lru.New(10)
	return
}

func userLogin(context *gin.Context) {
	request := struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Sub      struct {
		}
	}{}

	if err := context.BindJSON(&request); err != nil {
		context.Status(http.StatusBadRequest)
		return
	}

	// TODO 校验身份信息并设置用户权限

	response := User{
		UserID:      0,
		Username:    "root",
		AliasName:   "root",
		Permissions: []string{"*"},
	}
	session := uuid.NewString()
	loggedUser.Add(session, response)
	// 设置一小时的超时时间,
	context.SetCookie("session", session, 3600, "/", "", false, false)
	context.JSON(http.StatusOK, response)
}

func userLogout(context *gin.Context) {
	session, _ := context.Cookie("session")
	loggedUser.Remove(session)

	context.SetCookie("session", session, -1, "/", "", false, false)
	context.Status(http.StatusUnauthorized)
}

func userInsert(context *gin.Context) {
}

func userDelete(context *gin.Context) {
}

func userUpdate(context *gin.Context) {
}

func userQuery(context *gin.Context) {
}
