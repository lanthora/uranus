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
	UserID      uint64 `json:"userID" binding:"required"`
	Username    string `json:"username" binding:"required"`
	AliasName   string `json:"aliasName" binding:"required"`
	Permissions string `json:"permissions" binding:"required"`
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
			render.Status(context, render.StatusNotLoggedIn)
			context.Abort()
			return
		}

		user, ok := loggedUser.Get(session)
		if !ok {
			render.Status(context, render.StatusNotLoggedIn)
			context.Abort()
			return
		}

		g, err := glob.Compile(user.(User).Permissions)
		if err != nil {
			render.Status(context, render.StatusPermissionDenied)
			context.Abort()
			return
		}
		if !g.Match(context.Request.URL.Path) {
			render.Status(context, render.StatusPermissionDenied)
			context.Abort()
			return
		}
		context.Next()
	}
}

func userLogin(context *gin.Context) {
	request := struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	// TODO: 校验身份信息并设置用户权限

	response := User{
		UserID:      0,
		Username:    "root",
		AliasName:   "root",
		Permissions: "{*}",
	}
	session := uuid.NewString()
	loggedUser.Add(session, response)
	context.SetSameSite(http.SameSiteStrictMode)
	context.SetCookie("session", session, 0, "/", "", false, false)
	render.Status(context, render.StatusSuccess)
}

func userAlive(context *gin.Context) {
	render.Status(context, render.StatusSuccess)
}

func userInfo(context *gin.Context) {
	session, err := context.Cookie("session")
	if err != nil {
		render.Status(context, render.StatusNotLoggedIn)
		return
	}

	current, ok := loggedUser.Get(session)
	if !ok {
		render.Status(context, render.StatusNotLoggedIn)
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
