// Package handler 实现 HTTP 处理器层。
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/middleware"
	"github.com/onlineexam/onlineexam/internal/util"
)

// Success 统一成功响应：{ code:0, message:"ok", data:... }。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": constants.MsgOK, "data": data})
}

// SuccessMessage 带自定义 message 的成功响应。
func SuccessMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": message, "data": data})
}

// PageResult 统一分页响应。
func PageResult(c *gin.Context, list interface{}, total, page, pageSize int64) {
	Success(c, dto.PageResult{
		List:      list,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: util.TotalPages(total, pageSize),
	})
}

// Error 统一错误响应：将 AppError 或未知错误转换为 JSON。
func Error(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		status := httpStatusFromCode(appErr.Code)
		c.AbortWithStatusJSON(status, gin.H{"code": appErr.Code, "message": appErr.Message, "data": nil})
		return
	}
	util.Logger.Error("未包装错误", "path", c.FullPath(), "request_id", middleware.GetRequestID(c), "error", err.Error())
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"code":    constants.CodeInternalError,
		"message": constants.ErrorCodeText(constants.CodeInternalError),
		"data":    nil,
	})
}

func httpStatusFromCode(code int) int {
	switch code {
	case constants.CodeBadRequest, constants.CodeValidationFailed, constants.CodeQuestionTypeErr,
		constants.CodeUserInvalidRole, constants.CodeExamStatusErr, constants.CodeExamNotInWindow,
		constants.CodeRecordStatusErr, constants.CodeRecordExpired:
		return http.StatusBadRequest
	case constants.CodeUnauthorized:
		return http.StatusUnauthorized
	case constants.CodeForbidden, constants.CodeRecordNotOwner:
		return http.StatusForbidden
	case constants.CodeNotFound, constants.CodeUserNotFound, constants.CodeQuestionNotFound,
		constants.CodeExamNotFound, constants.CodeRecordNotFound, constants.CodeWrongBookNotFound,
		constants.CodeAuditNotFound:
		return http.StatusNotFound
	case constants.CodeConflict, constants.CodeDuplicateKey, constants.CodeUserEmailExists,
		constants.CodeRecordAlreadyDone, constants.CodeRecordStaleSave, constants.CodeWrongBookExists:
		return http.StatusConflict
	case constants.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
