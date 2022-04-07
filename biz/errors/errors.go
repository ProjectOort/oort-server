package errors

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type bizError struct {
	statusCode int
	msg        string
}

func New() *bizError {
	return &bizError{
		statusCode: http.StatusInternalServerError,
	}
}

func (e *bizError) StatusCode(code int) *bizError {
	e.statusCode = code
	return e
}

func (e *bizError) GetStatusCode() int {
	return e.statusCode
}

func (e *bizError) Msg(msg string) *bizError {
	e.msg = msg
	return e
}

func (e *bizError) GetMsg() string {
	return e.msg
}

func (e *bizError) Error() string {
	return fmt.Sprintf("msg: %s", e.msg)
}

func (e *bizError) WrapSelf() error {
	return errors.WithStack(e)
}

func As(err error) (*bizError, bool) {
	e := new(bizError)
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
