package logger

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

type Logger struct {
	*slog.Logger
}

func NewLogger() *Logger {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &Logger{logger}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}
