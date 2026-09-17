package errors

import "net/http"

// 业务错误码集中维护，禁止在业务代码里抛裸字符串。
const (
	CodeValidation       = "VALIDATION_FAILED"
	CodeNotFound         = "APPOINTMENT_NOT_FOUND"
	CodeForbidden        = "FORBIDDEN"
	CodeNotParticipant   = "NOT_APPOINTMENT_PARTICIPANT"
	CodeStatusConflict   = "APPOINTMENT_STATUS_CONFLICT"
	CodeTerminalState    = "APPOINTMENT_TERMINAL_STATE"
	CodeSlotUnavailable  = "TIME_SLOT_UNAVAILABLE"
	CodeCancelNotPending = "CANCEL_REQUEST_NOT_PENDING"
	CodeDuplicateActor   = "DUPLICATE_PARTICIPANT"
)

type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e BusinessError) Error() string { return e.Message }

func newErr(code, message string, status int) BusinessError {
	return BusinessError{Code: code, Message: message, Status: status}
}

func ValidationError(message string) BusinessError {
	return newErr(CodeValidation, message, http.StatusBadRequest)
}

func NotFoundError() BusinessError {
	return newErr(CodeNotFound, "交换预约不存在或已被清理", http.StatusNotFound)
}

func ForbiddenError(message string) BusinessError {
	return newErr(CodeForbidden, message, http.StatusForbidden)
}

func NotParticipantError() BusinessError {
	return newErr(CodeNotParticipant, "只有预约双方可以操作该预约", http.StatusForbidden)
}

func StatusConflictError(message string) BusinessError {
	return newErr(CodeStatusConflict, message, http.StatusConflict)
}

func TerminalStateError(message string) BusinessError {
	return newErr(CodeTerminalState, message, http.StatusConflict)
}

func SlotUnavailableError(message string) BusinessError {
	return newErr(CodeSlotUnavailable, message, http.StatusConflict)
}

func CancelNotPendingError(message string) BusinessError {
	return newErr(CodeCancelNotPending, message, http.StatusConflict)
}
