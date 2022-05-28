// SPDX-License-Identifier: AGPL-3.0-or-later
package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoute(engine *gin.Engine) {
	engine.POST("/user/login", userLogin)
	engine.POST("/user/logout", userLogout)
	engine.POST("/user/insert", userInsert)
	engine.POST("/user/delete", userDelete)
	engine.POST("/user/update", userUpdate)
	engine.POST("/user/query", userQuery)
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

	// TODO 校验身份信息

	response := struct {
		UserID    uint64 `json:"userID" binding:"required"`
		Username  string `json:"username" binding:"required"`
		AliasName string `json:"aliasName" binding:"required"`
		// 可以访问的路由 (wildcards)
		Permissions []string `json:"permissions" binding:"required"`
	}{
		UserID:      0,
		Username:    "root",
		AliasName:   "root",
		Permissions: []string{"*"},
	}

	context.JSON(http.StatusOK, response)
}

func userLogout(context *gin.Context) {
}

func userInsert(context *gin.Context) {
}

func userDelete(context *gin.Context) {
}

func userUpdate(context *gin.Context) {
}

func userQuery(context *gin.Context) {
}
