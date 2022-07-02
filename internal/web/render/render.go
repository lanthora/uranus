// SPDX-License-Identifier: AGPL-3.0-or-later
package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	StatusSuccess                        = 0
	StatusUnknownError                   = 1
	StatusNotLoggedIn                    = 2
	StatusPermissionDenied               = 3
	StatusInvalidArgument                = 4
	StatusLoginFaild                     = 5
	StatusCreateUserFailed               = 6
	StatusQuertUserFailed                = 7
	StatusUpdateUserFailed               = 8
	StatusDeleteUserFailed               = 9
	StatusUpdateProcessJudgeFailed       = 10
	StatusUpdateProcessEventStatusFailed = 11
	StatusQueryProcessEventFailed        = 12
	StatusProcessEnableFailed            = 13
	StatusProcessDisableFailed           = 14
	StatusAddFilePolicyConflict          = 15
	StatusAddFilePolicyFileNotExist      = 16
	StatusAddFilePolicyFailed            = 17
	StatusDeleteFilePolicyFailed         = 18
	StatusQueryFilePolicyListFailed      = 19
	StatusQueryFilePolicyByIdFailed      = 20
	StatusQueryFileEventListFailed       = 21
	StatusDeleteFileEventFailed          = 22
	StatusUpdateFilePolicyConflict       = 23
	StatusUpdateFilePolicyFileNotExist   = 24
	StatusUpdateFilePolicyFailed         = 25
	StatusFileEnableFailed               = 26
	StatusFileDisableFailed              = 27
	StatusProcessAutoTrustFailed         = 28
	StatusProcessGetAutoStatusFailed     = 29
)

var messages = [...]string{
	StatusSuccess:                        "success",
	StatusUnknownError:                   "unknown error",
	StatusNotLoggedIn:                    "not logged in",
	StatusPermissionDenied:               "permission denied",
	StatusInvalidArgument:                "invalid argument",
	StatusLoginFaild:                     "login failed",
	StatusCreateUserFailed:               "create user failed",
	StatusQuertUserFailed:                "query user failed",
	StatusUpdateUserFailed:               "update user failed",
	StatusDeleteUserFailed:               "delete user failed",
	StatusUpdateProcessJudgeFailed:       "update process judge failed",
	StatusUpdateProcessEventStatusFailed: "update process event status failed",
	StatusQueryProcessEventFailed:        "query process event failed",
	StatusProcessEnableFailed:            "process enable failed",
	StatusProcessDisableFailed:           "process disable failed",
	StatusAddFilePolicyConflict:          "add file policy conflict",
	StatusAddFilePolicyFileNotExist:      "add file policy file not exist",
	StatusAddFilePolicyFailed:            "add file policy failed",
	StatusDeleteFilePolicyFailed:         "delete file policy failed",
	StatusQueryFilePolicyListFailed:      "query file policy list failed",
	StatusQueryFilePolicyByIdFailed:      "query file policy by id failed",
	StatusQueryFileEventListFailed:       "query file event list failed",
	StatusDeleteFileEventFailed:          "delete file event failed",
	StatusUpdateFilePolicyConflict:       "update file policy conflict",
	StatusUpdateFilePolicyFileNotExist:   "update file policy file not exist",
	StatusUpdateFilePolicyFailed:         "update file policy failed",
	StatusFileEnableFailed:               "file enable failed",
	StatusFileDisableFailed:              "file disable failed",
	StatusProcessAutoTrustFailed:         "process auto trust failed",
	StatusProcessGetAutoStatusFailed:     "process get auto status failed",
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
