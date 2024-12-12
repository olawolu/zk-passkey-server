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
	loggerOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &loggerOpts))
	return &Logger{logger}
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
// 	// if l.LogLevel <= Silent {
// 	// 	return
// 	// }

// 		elapsed := time.Since(begin)
// 	switch {
// 	case err != nil && l.LogLevel >= Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
// 		sql, rows := fc()
// 		if rows == -1 {
// 			l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
// 		} else {
// 			l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
// 		}
// 	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
// 		sql, rows := fc()
// 		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
// 		if rows == -1 {
// 			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
// 		} else {
// 			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
// 		}
// 	case l.LogLevel == Info:
// 		sql, rows := fc()
// 		if rows == -1 {
// 			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
// 		} else {
// 			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
// 		}
// 	}
// }
