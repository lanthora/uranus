// SPDX-License-Identifier: AGPL-3.0-or-later
package user

import (
	"net/http"
	"uranus/internal/web/render"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/glob"
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
	engine.POST("/user/alive", userAlive)
	engine.POST("/user/info", userInfo)
	engine.POST("/user/logout", userLogout)
	engine.POST("/user/insert", userInsert)
	engine.POST("/user/delete", userDelete)
	engine.POST("/user/update", userUpdate)
	engine.POST("/user/query", userQuery)

	loggedUser, err = lru.New(10)
	return
}

func Middleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		if context.Request.Method == http.MethodGet {
			context.Next()
			return
		}
		if context.Request.URL.Path == "/user/login" {
			context.Next()
			return
		}
		session, err := context.Cookie("session")
		if err != nil {
			render.Status(context, render.StatusUnauthorized)
			context.Abort()
			return
		}

		user, ok := loggedUser.Get(session)
		if !ok {
			render.Status(context, render.StatusUnauthorized)
			context.Abort()
			return
		}

		for _, p := range user.(User).Permissions {
			g, err := glob.Compile(p)
			if err != nil {
				continue
			}
			if !g.Match(context.Request.URL.Path) {
				continue
			}
			context.Next()
			return
		}
		render.Status(context, render.StatusLocked)
		context.Abort()
	}
}

func userLogin(context *gin.Context) {
	request := struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Sub      struct {
		}
	}{}

	if err := context.BindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalid)
		return
	}

	// TODO: 校验身份信息并设置用户权限

	response := User{
		UserID:      0,
		Username:    "root",
		AliasName:   "root",
		Permissions: []string{"*"},
	}
	session := uuid.NewString()
	loggedUser.Add(session, response)
	// 设置一小时的超时时间
	context.SetSameSite(http.SameSiteStrictMode)
	context.SetCookie("session", session, 0, "/", "", false, false)
	render.Success(context, response)
}

func userAlive(context *gin.Context) {
	// 已经通过中间件校验,能走到这里说明已经登录成功
	render.Status(context, render.StatusSuccess)
}

func userInfo(context *gin.Context) {
	session, err := context.Cookie("session")
	if err != nil {
		render.Status(context, render.StatusUnauthorized)
		return
	}

	current, ok := loggedUser.Get(session)
	if !ok {
		render.Status(context, render.StatusUnauthorized)
		return
	}
	context.SetSameSite(http.SameSiteStrictMode)
	context.SetCookie("session", session, 0, "/", "", false, false)
	render.Success(context, current)
}

func userLogout(context *gin.Context) {
	session, _ := context.Cookie("session")
	loggedUser.Remove(session)

	context.SetSameSite(http.SameSiteStrictMode)
	context.SetCookie("session", session, -1, "/", "", false, false)
	render.Status(context, render.StatusSuccess)
}

func userInsert(context *gin.Context) {
}

func userDelete(context *gin.Context) {
}

func userUpdate(context *gin.Context) {
}

func userQuery(context *gin.Context) {
}
