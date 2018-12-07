package election

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// buildLogFields - build the log fields
func (e *ElectionManager) buildLogFields(function string) []zapcore.Field {

	return []zapcore.Field{
		zap.String("package", "election"),
		zap.String("func", function),
	}
}

// logError - logs the error message
func (e *ElectionManager) logError(function, message string) {

	e.logger.Error(message, e.buildLogFields(function)...)
}

// logInfo - logs the info message
func (e *ElectionManager) logInfo(function, message string) {

	e.logger.Info(message, e.buildLogFields(function)...)
}
