package cockroachdb

import (
	"github.com/uol/gobol"
)

const (
	recordNotFoundMsg string = "record not found"
	noErrors          uint8  = 0
	noRecords         uint8  = 1
	hasErrors         uint8  = 2
	cPackage          string = "cockroach"
)

// New - creates a new database error
func New(e error, msg, pkg, function, errorCode string, httpCode int) gobol.Error {
	return customError{
		e,
		msg,
		pkg,
		function,
		errorCode,
		httpCode,
	}
}

type customError struct {
	error
	msg       string
	pkg       string
	function  string
	errorCode string
	httpCode  int
}

func (e customError) Package() string {
	return e.pkg
}

func (e customError) Function() string {
	return e.function
}

func (e customError) Message() string {
	return e.msg
}

func (e customError) StatusCode() int {
	return e.httpCode
}

func (e customError) ErrorCode() string {
	return e.errorCode
}

func errBasic(function, msg, errorCode string, code int, e error) gobol.Error {
	if e != nil {
		return New(
			e,
			msg,
			cPackage,
			function,
			errorCode,
			code,
		)
	}
	return nil
}

// checkErrors - check if error list has a 'no record found' message or it has errors
func checkErrors(errors []error) uint8 {

	if errors != nil && len(errors) > 0 {

		for _, e := range errors {
			if e.Error() == recordNotFoundMsg {
				return noRecords
			}
		}

		return hasErrors
	}

	return noErrors
}
