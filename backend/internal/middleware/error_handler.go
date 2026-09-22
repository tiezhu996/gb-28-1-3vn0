package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/util"
)

// ErrorHandler 统一错误处理与 panic 恢复：将 panic 与业务错误转换为统一 JSON 响应。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				util.Logger.Error(fmt.Sprintf(constants.LogPanicRecovered, GetRequestID(c), c.FullPath(), r))
				writeError(c, http.StatusInternalServerError, constants.CodeInternalError, constants.ErrorCodeText(constants.CodeInternalError))
				c.Abort()
			}
		}()
		c.Next()
		// 若 handler 未写入响应且存在错误，统一包装（gin 的 c.Errors）
		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last().Err
			var appErr *util.AppError
			if errors.As(err, &appErr) {
				writeError(c, httpStatusFromCode(appErr.Code), appErr.Code, appErr.Message)
			} else {
				util.Logger.Error(fmt.Sprintf(constants.LogRequestFailed, GetRequestID(c), c.Request.Method, c.Request.URL.Path, err.Error()))
				writeError(c, http.StatusInternalServerError, constants.CodeInternalError, constants.ErrorCodeText(constants.CodeInternalError))
			}
		}
	}
}

func writeError(c *gin.Context, status, code int, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"code":    code,
		"message": message,
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
	case constants.CodeForbidden:
		return http.StatusForbidden
	case constants.CodeNotFound, constants.CodeUserNotFound, constants.CodeQuestionNotFound,
		constants.CodeExamNotFound, constants.CodeRecordNotFound, constants.CodeWrongBookNotFound,
		constants.CodeAuditNotFound:
		return http.StatusNotFound
	case constants.CodeConflict, constants.CodeDuplicateKey, constants.CodeUserEmailExists,
		constants.CodeRecordAlreadyDone, constants.CodeWrongBookExists:
		return http.StatusConflict
	case constants.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
