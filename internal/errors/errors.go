package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 通用错误
	ErrCodeUnknown ErrorCode = iota + 1000
	ErrCodeInvalidRequest
	ErrCodeNotFound
	ErrCodeUnauthorized
	ErrCodeForbidden
	ErrCodeInternalServer

	// 数据库错误
	ErrCodeDatabaseError ErrorCode = iota + 2000
	ErrCodeRecordNotFound
	ErrCodeDuplicateEntry

	// 认证错误
	ErrCodeInvalidCredentials ErrorCode = iota + 3000
	ErrCodeTokenExpired
	ErrCodeAPIKeyInvalid
	ErrCodeAPIKeyExpired

	// 配置错误
	ErrCodeConfigError ErrorCode = iota + 4000
	ErrCodeConfigNotFound
)

// AppError 应用错误
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidRequest:
		return http.StatusBadRequest
	case ErrCodeNotFound, ErrCodeRecordNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized, ErrCodeInvalidCredentials, ErrCodeTokenExpired, ErrCodeAPIKeyInvalid, ErrCodeAPIKeyExpired:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeInternalServer, ErrCodeDatabaseError, ErrCodeConfigError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// 错误构造函数
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 常用错误
var (
	ErrInvalidRequest   = New(ErrCodeInvalidRequest, "无效的请求")
	ErrNotFound         = New(ErrCodeNotFound, "资源未找到")
	ErrUnauthorized     = New(ErrCodeUnauthorized, "未授权")
	ErrForbidden        = New(ErrCodeForbidden, "禁止访问")
	ErrInternalServer   = New(ErrCodeInternalServer, "服务器内部错误")
	ErrDatabaseError    = New(ErrCodeDatabaseError, "数据库错误")
	ErrRecordNotFound   = New(ErrCodeRecordNotFound, "记录未找到")
	ErrDuplicateEntry   = New(ErrCodeDuplicateEntry, "记录已存在")
	ErrInvalidCredentials = New(ErrCodeInvalidCredentials, "无效的凭据")
	ErrTokenExpired     = New(ErrCodeTokenExpired, "令牌已过期")
	ErrAPIKeyInvalid    = New(ErrCodeAPIKeyInvalid, "无效的 API Key")
	ErrAPIKeyExpired    = New(ErrCodeAPIKeyExpired, "API Key 已过期")
	ErrConfigError      = New(ErrCodeConfigError, "配置错误")
	ErrConfigNotFound   = New(ErrCodeConfigNotFound, "配置文件未找到")
)

