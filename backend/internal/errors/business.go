package errors

import "net/http"

// BusinessError 业务异常：错误码与错误消息集中管理，Status 用于 HTTP 映射。
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e BusinessError) Error() string { return e.Message }

// New 构造业务异常。
func New(status int, code, message string) BusinessError {
	return BusinessError{Code: code, Message: message, Status: status}
}

// BadRequest 构造 400 业务异常。
func BadRequest(code, message string) BusinessError {
	return New(http.StatusBadRequest, code, message)
}

// Forbidden 构造 403 业务异常。
func Forbidden(code, message string) BusinessError {
	return New(http.StatusForbidden, code, message)
}

// NotFound 构造 404 业务异常。
func NotFound(code, message string) BusinessError {
	return New(http.StatusNotFound, code, message)
}

// Conflict 构造 409 业务异常（状态冲突，如终态不可变更、确认与撤回竞争失败）。
func Conflict(code, message string) BusinessError {
	return New(http.StatusConflict, code, message)
}
