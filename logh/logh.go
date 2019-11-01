package logh

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

//
// Has some useful logging functions.
// logh -> log helper
// @author rnojiri
//

// Level - type
type Level string

const (
	// INFO - log level
	INFO Level = "info"

	// DEBUG - log level
	DEBUG Level = "debug"

	// WARNING - log level
	WARNING Level = "warning"

	// ERROR - log level
	ERROR Level = "error"

	// FATAL - log level
	FATAL Level = "fatal"

	// PANIC - log level
	PANIC Level = "panic"

	// NONE - log level
	NONE Level = "none"

	// SILENT - log level
	SILENT Level = "silent"
)

// Format - the logger's output format
type Format string

const (
	// JSON - json format
	JSON Format = "json"

	// CONSOLE - plain text format
	CONSOLE Format = "console"
)

var (
	stdout zerolog.Logger
	stderr zerolog.Logger
)

// EventLoggers - a struct containing all valid event loggers (each one can be null if not enabled)
type EventLoggers struct {
	Info    *zerolog.Event
	Debug   *zerolog.Event
	Warning *zerolog.Event
	Error   *zerolog.Event
	Fatal   *zerolog.Event
	Panic   *zerolog.Event
}

// ConfigureGlobalLogger - configures the logger globally
func ConfigureGlobalLogger(lvl Level, fmt Format) {

	switch lvl {
	case INFO:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case DEBUG:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case WARNING:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case ERROR:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case PANIC:
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case FATAL:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case NONE:
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	case SILENT:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	var out io.Writer
	var err io.Writer

	if fmt == CONSOLE {
		out = zerolog.ConsoleWriter{Out: os.Stdout}
		err = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		out = os.Stdout
		err = os.Stderr
	}

	stdout = zerolog.New(out).With().Timestamp().Logger()
	stderr = zerolog.New(err).With().Timestamp().Logger()
}

// SendToStdout - logs a output with no log format
func SendToStdout(output string) {

	fmt.Println(output)
}

// InfoLogger - returns the info event logger if any
func InfoLogger() *zerolog.Event {
	if e := stdout.Info(); e.Enabled() {
		return e
	}
	return nil
}

// DebugLogger - returns the debug event logger if any
func DebugLogger() *zerolog.Event {
	if e := stdout.Debug(); e.Enabled() {
		return e
	}
	return nil
}

// WarningLogger - returns the error event logger if any
func WarningLogger() *zerolog.Event {
	if e := stdout.Warn(); e.Enabled() {
		return e
	}
	return nil
}

// ErrorLogger - returns the error event logger if any
func ErrorLogger() *zerolog.Event {
	if e := stderr.Error(); e.Enabled() {
		return e
	}
	return nil
}

// PanicLogger - returns the error event logger if any
func PanicLogger() *zerolog.Event {
	if e := stderr.Panic(); e.Enabled() {
		return e
	}
	return nil
}

// FatalLogger - returns the error event logger if any
func FatalLogger() *zerolog.Event {
	if e := stderr.Fatal(); e.Enabled() {
		return e
	}
	return nil
}

// CreateContexts - creates loggers with context
func CreateContexts(incInfo, incDebug, incWarning, incError, incFatal, incPanic bool, keyValues ...string) *EventLoggers {

	numKeyValues := len(keyValues)
	if numKeyValues%2 != 0 {
		panic("the number of arguments must be even")
	}

	el := &EventLoggers{}

	if incInfo {
		el.Info = addContext(InfoLogger(), numKeyValues, keyValues)
	}

	if incDebug {
		el.Debug = addContext(DebugLogger(), numKeyValues, keyValues)
	}

	if incWarning {
		el.Warning = addContext(WarningLogger(), numKeyValues, keyValues)
	}

	if incError {
		el.Error = addContext(ErrorLogger(), numKeyValues, keyValues)
	}

	if incFatal {
		el.Fatal = addContext(FatalLogger(), numKeyValues, keyValues)
	}

	if incPanic {
		el.Panic = addContext(PanicLogger(), numKeyValues, keyValues)
	}

	return el
}

// addContext - add event logger context
func addContext(eventlLogger *zerolog.Event, numKeyValues int, keyValues []string) *zerolog.Event {

	if eventlLogger == nil {
		return nil
	}

	for j := 0; j < numKeyValues; j += 2 {
		eventlLogger = eventlLogger.Str(keyValues[j], keyValues[j+1])
	}

	return eventlLogger
}
