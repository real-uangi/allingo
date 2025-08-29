package business

import (
	"bytes"
	"github.com/real-uangi/allingo/common/trace"
	"net/http"
)

type RuntimeError struct {
	msg  string
	code int
}

func msgWithStack(msg string) string {
	buffer := new(bytes.Buffer)
	buffer.WriteString(msg)
	buffer.WriteRune('\n')
	buffer.WriteString(trace.Stack(3))
	return buffer.String()
}

func NewError(msg string) *RuntimeError {
	return &RuntimeError{
		msg:  msgWithStack(msg),
		code: http.StatusInternalServerError,
	}
}

func NewErrorWithCode(msg string, code int) *RuntimeError {
	return &RuntimeError{
		msg:  msgWithStack(msg),
		code: code,
	}
}

func NewErrorWithStatus(status int) error {
	return &RuntimeError{
		msg:  msgWithStack(http.StatusText(status)),
		code: status,
	}
}

func preDefine(msg string, code int) *RuntimeError {
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
	ErrPermissionDenied = preDefine("Permission Denied", http.StatusForbidden)
	ErrForbidden        = preDefine("Forbidden", http.StatusForbidden)
	ErrUnauthorized     = preDefine("Unauthorized", http.StatusUnauthorized)
	ErrNotFound         = preDefine("Not Found", http.StatusNotFound)
	ErrBadRequest       = preDefine("Bad Request", http.StatusBadRequest)
	ErrDataCheckError   = preDefine("Data Check Error", http.StatusUnprocessableEntity)
	ErrTooManyRequests  = preDefine("Too Many Requests", http.StatusTooManyRequests)
	ErrBadGateway       = preDefine("Bad Gateway", http.StatusBadGateway)
	ErrIllegalUUID      = preDefine("Illegal ID", http.StatusBadRequest)
)
