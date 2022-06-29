// SPDX-License-Identifier: AGPL-3.0-or-later
package user

import (
	"database/sql"
	"net/http"
	"uranus/internal/web/render"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/glob"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
)

type Worker struct {
	engine         *gin.Engine
	dataSourceName string

	onlineUserMax int
	loggedUser    *lru.Cache
}

func Init(engine *gin.Engine, dataSourceName string) (err error) {
	onlineUserMax := 10
	loggedUser, err := lru.New(onlineUserMax)
	if err != nil {
		return
	}

	w := &Worker{
		loggedUser:     loggedUser,
		onlineUserMax:  onlineUserMax,
		engine:         engine,
		dataSourceName: dataSourceName,
	}

	w.engine.Use(w.middleware())

	w.engine.POST("/user/login", w.userLogin)
	w.engine.POST("/user/alive", w.userAlive)
	w.engine.POST("/user/info", w.userInfo)
	w.engine.POST("/user/logout", w.userLogout)
	w.engine.POST("/user/insert", w.userInsert)
	w.engine.POST("/user/delete", w.userDelete)
	w.engine.POST("/user/update", w.userUpdate)
	w.engine.POST("/user/query", w.userQuery)

	if err = w.initUserTable(); err != nil {
		return
	}
	return

}

type User struct {
	UserID      uint64 `json:"userID" binding:"required"`
	Username    string `json:"username" binding:"required"`
	AliasName   string `json:"aliasName" binding:"required"`
	Permissions string `json:"permissions" binding:"required"`
}

func (w *Worker) middleware() gin.HandlerFunc {
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

		user, ok := w.loggedUser.Get(session)
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

func (w *Worker) userLogin(context *gin.Context) {
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
		render.Status(context, render.StatusLoginFaild)
		return
	}

	if err != nil {
		render.Status(context, render.StatusUnknownError)
		return
	}

	if !ok {
		render.Status(context, render.StatusLoginFaild)
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

func (w *Worker) userAlive(context *gin.Context) {
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) userInfo(context *gin.Context) {
	session, err := context.Cookie("session")
	if err != nil {
		render.Status(context, render.StatusNotLoggedIn)
		return
	}

	current, ok := w.loggedUser.Get(session)
	if !ok {
		render.Status(context, render.StatusNotLoggedIn)
		return
	}
	updateSession(context, session)
	render.Success(context, current)
}

func (w *Worker) userLogout(context *gin.Context) {
	session, _ := context.Cookie("session")
	w.loggedUser.Remove(session)

	deleteSession(context)
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) userInsert(context *gin.Context) {
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
		render.Status(context, render.StatusCreateUserFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) userQuery(context *gin.Context) {
	users, err := w.queryAllUser()
	if err != nil {
		render.Status(context, render.StatusQuertUserFailed)
		return
	}
	render.Success(context, users)
}

func (w *Worker) userDelete(context *gin.Context) {
	request := struct {
		UserID uint64 `json:"userID" binding:"number"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}
	if ok := w.deleteUser(request.UserID); !ok {
		render.Status(context, render.StatusDeleteUserFailed)
		return
	}
	render.Status(context, render.StatusSuccess)
}

func (w *Worker) userUpdate(context *gin.Context) {
	request := struct {
		UserID      uint64 `json:"userID" binding:"number"`
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		AliasName   string `json:"aliasName" binding:"required"`
		Permissions string `json:"permissions" binding:"required"`
	}{}

	if err := context.ShouldBindJSON(&request); err != nil {
		render.Status(context, render.StatusInvalidArgument)
		return
	}

	if ok := w.updateUserInfo(request.UserID, request.Username, request.Password, request.AliasName, request.Permissions); !ok {
		render.Status(context, render.StatusCreateUserFailed)
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
