package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	StatusSuccess      = 0
	StatusUnknown      = 1
	StatusUnauthorized = 2
	StatusLocked       = 3
	StatusInvalid      = 4
)

func Success(context *gin.Context, data interface{}) {
	response := struct {
		Status int         `json:"status" binding:"required"`
		Data   interface{} `json:"data" binding:"required"`
	}{
		Status: StatusSuccess,
		Data:   data,
	}
	context.JSON(http.StatusOK, response)
}

func Status(context *gin.Context, status int) {
	response := struct {
		Status int `json:"status" binding:"required"`
	}{
		Status: status,
	}
	context.JSON(http.StatusOK, response)
}
