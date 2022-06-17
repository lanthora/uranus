package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	StatusSuccess          = 0
	StatusUnknownError     = 1
	StatusNotLoggedIn      = 2
	StatusPermissionDenied = 3
	StatusInvalidArgument  = 4
	StatusLoginFaild       = 5
	StatusCreateUserFailed = 6
	StatusQuertUserFailed  = 7
	StatusUpdateUserFailed = 8
	StatusDeleteUserFailed = 9
)

var messages = [...]string{
	StatusSuccess:          "success",
	StatusUnknownError:     "unknown error",
	StatusNotLoggedIn:      "not logged in",
	StatusPermissionDenied: "permission denied",
	StatusInvalidArgument:  "invalid argument",
	StatusLoginFaild:       "login failed",
	StatusCreateUserFailed: "create user failed",
	StatusQuertUserFailed:  "query user failed",
	StatusUpdateUserFailed: "update user failed",
	StatusDeleteUserFailed: "delete user failed",
}

func Success(context *gin.Context, data interface{}) {
	response := struct {
		Status  int         `json:"status" binding:"required"`
		Message string      `json:"message" binding:"required"`
		Data    interface{} `json:"data" binding:"required"`
	}{
		Status:  StatusSuccess,
		Message: messages[StatusSuccess],
		Data:    data,
	}
	context.JSON(http.StatusOK, response)
}

func Status(context *gin.Context, status int) {
	response := struct {
		Status  int    `json:"status" binding:"required"`
		Message string `json:"message" binding:"required"`
	}{
		Status:  status,
		Message: messages[status],
	}
	context.JSON(http.StatusOK, response)
}
