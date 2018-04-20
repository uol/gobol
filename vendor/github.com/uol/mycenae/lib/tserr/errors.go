package tserr

import (
	"github.com/uol/gobol"
)

func New(e error, msg string, httpCode int, lf map[string]interface{}) gobol.Error {
	return customError{
		e,
		msg,
		httpCode,
		lf,
	}
}

type customError struct {
	error
	msg      string
	httpCode int
	lf       map[string]interface{}
}

func (e customError) Message() string {
	return e.msg
}

func (e customError) StatusCode() int {
	return e.httpCode
}

func (e customError) LogFields() map[string]interface{} {
	return e.lf
}
