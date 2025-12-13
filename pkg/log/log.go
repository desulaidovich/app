package log

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	TimeFormat = "02-01-2006 15:04:05.000"
)

type Logger struct {
	logger *slog.Logger
}

func New(level string) *Logger {
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"err":   slog.LevelError,
	}

	var (
		lvl   slog.Level
		exist bool
	)

	lvl, exist = levelMap[strings.ToLower(level)]
	if !exist {
		lvl = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				return slog.String(slog.TimeKey, time.Now().Format(TimeFormat))

			case slog.LevelKey:
				level := map[string]string{
					"DEBUG": "DBG",
					"INFO":  "INF",
					"WARN":  "WRN",
					"ERROR": "ERR",
				}

				return slog.String(slog.LevelKey, level[a.Value.String()])
			}

			return a
		},
	}

	return &Logger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, opts)),
	}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *Logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Error(msg)
}
