package business

import (
	"bytes"
	"fmt"
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
		msg:  msg,
		code: http.StatusInternalServerError,
	}
}

func NewStackError(msg string) *RuntimeError {
	return &RuntimeError{
		msg:  msgWithStack(msg),
		code: http.StatusInternalServerError,
	}
}

func NewErrorf(format string, args ...interface{}) *RuntimeError {
	return &RuntimeError{
		msg:  fmt.Sprintf(format, args...),
		code: http.StatusInternalServerError,
	}
}

func NewStackErrorf(format string, args ...interface{}) *RuntimeError {
	return &RuntimeError{
		msg:  msgWithStack(fmt.Sprintf(format, args...)),
		code: http.StatusInternalServerError,
	}
}

func NewErrorWithCode(msg string, code int) *RuntimeError {
	return &RuntimeError{
		msg:  msg,
		code: code,
	}
}

func NewStackErrorWithCode(msg string, code int) *RuntimeError {
	return &RuntimeError{
		msg:  msgWithStack(msg),
		code: code,
	}
}

func NewErrorWithStatus(status int) error {
	return &RuntimeError{
		msg:  http.StatusText(status),
		code: status,
	}
}

func NewStackErrorWithStatus(status int) error {
	return &RuntimeError{
		msg:  msgWithStack(http.StatusText(status)),
		code: status,
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
