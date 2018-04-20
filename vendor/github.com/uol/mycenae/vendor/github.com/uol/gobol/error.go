package gobol

type Error interface {
	error
	StatusCode() int
	Message() string
	LogFields() map[string]interface{}
}
