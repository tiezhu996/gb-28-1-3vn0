package util

import "fmt"

// AppError 统一业务错误，包含错误码与可读信息。
// 由 handler 层转换为统一 JSON 响应；service 层必须使用 %w 包裹底层错误以保留错误链。
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s cause=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is / errors.As 错误链。
func (e *AppError) Unwrap() error { return e.Err }

// NewAppError 构造业务错误。
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WrapAppError 包裹底层错误并保留业务错误码。
func WrapAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
