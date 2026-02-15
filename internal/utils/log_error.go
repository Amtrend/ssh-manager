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
func LogErrorf(format string, args ...interface{}) {
	logger.Error("Application Error", "details", args)
}

// LogInfo log for information.
func LogInfo(msg string, keyValues ...interface{}) {
	logger.Info(msg, keyValues...)
}
