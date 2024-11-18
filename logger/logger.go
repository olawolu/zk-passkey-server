package logger

import (
	"context"
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

func (l *Logger) Debug(ctx context.Context, msg string) {
	fv := ctx.Value("attrs")
	l.DebugContext(ctx, msg, fv)
}

func (l *Logger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	fv := ctx.Value("attrs")
	l.InfoContext(ctx, msg, fv)
}

func (l *Logger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	fv := ctx.Value("attrs")
	l.ErrorContext(ctx, msg, fv)
}
