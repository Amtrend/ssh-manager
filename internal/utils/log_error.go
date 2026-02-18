package utils

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

// InitLogger initializes the logger for logs.
func InitLogger() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// LogErrorf error log.
func LogErrorf(msg string, err error, keysAndValues ...interface{}) {
	if logger == nil {
		InitLogger()
	}

	args := append([]interface{}{"error", err}, keysAndValues...)
	logger.Error(msg, args...)
}

// LogInfo log for information.
func LogInfo(msg string, keyValues ...interface{}) {
	if logger == nil {
		InitLogger()
	}
	logger.Info(msg, keyValues...)
}
