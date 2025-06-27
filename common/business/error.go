package business

import "net/http"

type RuntimeError struct {
	msg  string
	code int
}

func NewError(msg string) *RuntimeError {
	return &RuntimeError{
		msg:  msg,
		code: http.StatusInternalServerError,
	}
}

func NewErrorWithCode(msg string, code int) *RuntimeError {
	return &RuntimeError{
		msg:  msg,
		code: code,
	}
}

func NewBadRequest(msg string) *RuntimeError {
	return NewErrorWithCode(msg, http.StatusBadRequest)
}

func (e *RuntimeError) Error() string {
	return e.msg
}

func (e *RuntimeError) GetStatusCode() int {
	return e.code
}

var (
	ErrPermissionDenied = NewErrorWithCode("Permission Denied", http.StatusForbidden)
	ErrForbidden        = NewErrorWithCode("Forbidden", http.StatusForbidden)
	ErrUnauthorized     = NewErrorWithCode("Unauthorized", http.StatusUnauthorized)
	ErrNotFound         = NewErrorWithCode("Not Found", http.StatusNotFound)
	ErrBadRequest       = NewErrorWithCode("Bad Request", http.StatusBadRequest)
	ErrDataCheckError   = NewErrorWithCode("Data Check Error", http.StatusUnprocessableEntity)
	ErrTooManyRequests  = NewErrorWithCode("Too Many Requests", http.StatusTooManyRequests)
	ErrBadGateway       = NewErrorWithCode("Bad Gateway", http.StatusBadGateway)
	ErrIllegalUUID      = NewErrorWithCode("Illegal ID", http.StatusBadRequest)
)
