// SPDX-License-Identifier: AGPL-3.0-or-later
package user

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/glob"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lanthora/uranus/internal/web/render"
)

type Worker struct {
	engine *gin.Engine
	db     *sql.DB

	onlineUserMax int
	loggedUser    *lru.Cache
}

func Init(engine *gin.Engine, db *sql.DB) (err error) {
	onlineUserMax := 10
	loggedUser, err := lru.New(onlineUserMax)
	if err != nil {
		return
	}

	w := &Worker{
		loggedUser:    loggedUser,
		onlineUserMax: onlineUserMax,
		engine:        engine,
		db:            db,
	}

	w.engine.Use(w.middleware())

	w.engine.POST("/auth/login", w.login)
	w.engine.POST("/auth/showCurrentUserInfo", w.showCurrentUserInfo)
	w.engine.POST("/auth/logout", w.logout)

	w.engine.POST("/admin/addUser", w.addUser)
	w.engine.POST("/admin/deleteUser", w.deleteUser)
	w.engine.POST("/admin/updateUserInfo", w.updateUserInfo)
	w.engine.POST("/admin/listAllUsers", w.listAllUsers)

	if err = w.initUserTable(); err != nil {
		return
	}
	return

}

type User struct {
	UserID      int64  `json:"userID"`
	Username    string `json:"username"`
	AliasName   string `json:"aliasName"`
	Permissions string `json:"permissions"`
}

func (w *Worker) middleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 静态资源不校验权限
		staticFiles := map[string]bool{"": true, "/": true, "/favicon.ico": true, "/index.html": true, "/asset-manifest.json": true}
		if staticFiles[context.Request.URL.Path] {
			context.Next()
			return
		}
		if strings.HasPrefix(context.Request.URL.Path, "/static/") {
			context.Next()
			return
		}

		// pprof 调试相关的校验由 ctrl 中的 middleware 校验
		if strings.HasPrefix(context.Request.URL.Path, pprof.DefaultPrefix) {
			context.Next()
			return
		}

		// 不校验登录接口
		if context.Request.URL.Path == "/auth/login" {
			context.Next()
			return
		}
		session, err := context.Cookie("session")
		if err != nil {
			render.Status(context, render.StatusUserNotLoggedIn)
			context.Abort()
			return
		}

		user, ok := w.loggedUser.Get(session)
		if !ok {
			render.Status(context, render.StatusUserNotLoggedIn)
			context.Abort()
			return
		}

		g, err := glob.Compile(user.(User).Permissions)
		if err != nil {
			render.Status(context, render.StatusUserPermissionDenied)
			context.Abort()
			return
		}
		if !g.Match(context.Request.URL.Path) {
			render.Status(context, render.StatusUserPermissionDenied)
			context.Abort()
			return
		}
		context.Next()
	}
}

func (w *Worker) login(context *gin.Context) {
	request := struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}{}
	deleteSession(context)

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if w.noUser() {
		w.createUser(request.Username, request.Password, request.Username, `{*}`)
	}

	ok, err := w.checkUserPassword(request.Username, request.Password)
	if err == sql.ErrNoRows {
		render.Status(context, render.StatusUserLoginFaild)
		return
	}

	if err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}

	if !ok {
		render.Status(context, render.StatusUserLoginFaild)
		return
	}

	response, err := w.queryUserByUsername(request.Username)
	if err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}
	session := uuid.NewString()
	w.loggedUser.Add(session, response)
	updateSession(context, session)
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) showCurrentUserInfo(context *gin.Context) {
	session, err := context.Cookie("session")
	if err != nil {
		render.Status(context, render.StatusUserNotLoggedIn)
		return
	}

	current, ok := w.loggedUser.Get(session)
	if !ok {
		render.Status(context, render.StatusUserNotLoggedIn)
		return
	}
	updateSession(context, session)
	render.Success(context, current)
}

func (w *Worker) logout(context *gin.Context) {
	session, _ := context.Cookie("session")
	w.loggedUser.Remove(session)

	deleteSession(context)
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) addUser(context *gin.Context) {
	request := struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		AliasName   string `json:"aliasName" binding:"required"`
		Permissions string `json:"permissions" binding:"required"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if err := w.createUser(request.Username, request.Password, request.AliasName, request.Permissions); err != nil {
		render.Status(context, render.StatusUserCreateUserFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) listAllUsers(context *gin.Context) {
	users, err := w.queryAllUser()
	if err != nil {
		render.Status(context, render.StatusUserQueryUserFailed)
		return
	}
	render.Success(context, users)
}

func (w *Worker) deleteUser(context *gin.Context) {
	request := struct {
		UserID int64 `json:"userID" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}
	if ok := w.deleteUserByID(request.UserID); !ok {
		render.Status(context, render.StatusUserDeleteUserFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) updateUserInfo(context *gin.Context) {
	request := struct {
		UserID      int64  `json:"userID" binding:"number"`
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		AliasName   string `json:"aliasName" binding:"required"`
		Permissions string `json:"permissions" binding:"required"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if ok := w.updateUserInfoByID(request.UserID, request.Username, request.Password, request.AliasName, request.Permissions); !ok {
		render.Status(context, render.StatusUserUpdateUserFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func updateSession(context *gin.Context, session string) {
	context.SetSameSite(http.SameSiteStrictMode)
	context.SetCookie("session", session, 0, "/", "", false, false)
}

func deleteSession(context *gin.Context) {
	context.SetSameSite(http.SameSiteStrictMode)
	context.SetCookie("session", "deleted", -1, "/", "", false, false)
}
