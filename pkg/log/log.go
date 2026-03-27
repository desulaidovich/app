package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"
)

const (
	OutputJSON = "json"
	OutputText = "text"

	LevelPanic = slog.Level(12)
)

var levelNames = map[slog.Level]string{
	slog.LevelDebug: "DBG",
	slog.LevelInfo:  "INF",
	slog.LevelWarn:  "WRN",
	slog.LevelError: "ERR",
	LevelPanic:      "PNC",
}

// Logger определяет интерфейс для логирования с поддержкой уровней и структурированных полей.
type Logger interface {
	// Debug логирует сообщение на уровне DEBUG.
	Debug(msg string, args ...any)
	// Info логирует сообщение на уровне INFO.
	Info(msg string, args ...any)
	// Warn логирует сообщение на уровне WARN.
	Warn(msg string, args ...any)
	// Error логирует сообщение на уровне ERROR.
	Error(msg string, args ...any)
	// Panic логирует сообщение на уровне PANIC и завершает программу с кодом 1.
	Panic(msg string, args ...any)
	// With возвращает новый логгер с добавленными полями, которые будут включены во все последующие записи.
	With(fields map[string]any) Logger
}

type loggerImpl struct {
	logger *slog.Logger
}

type options struct {
	level      slog.Level
	output     io.Writer
	format     string
	timeFormat string
}

type Option func(*options) error

// WithLevel устанавливает уровень логирования.
// Допустимые значения: debug, info, warn, error.
func WithLevel(level string) Option {
	m := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	return func(o *options) error {
		l, ok := m[strings.ToLower(level)]
		if !ok {
			return fmt.Errorf("unsupported log level: %q (valid: debug, info, warn, error)", level)
		}
		o.level = l
		return nil
	}
}

// WithOutput устанавливает вывод логов. По умолчанию используется os.Stdout.
func WithOutput(w io.Writer) Option {
	return func(o *options) error {
		if w == nil {
			return errors.New("output writer cannot be nil")
		}
		o.output = w
		return nil
	}
}

// WithFormat устанавливает формат вывода логов.
// Допустимые значения: json, text.
func WithFormat(format string) Option {
	return func(o *options) error {
		if format != OutputJSON && format != OutputText {
			return fmt.Errorf("unsupported log format: %q (valid: %s, %s)", format, OutputJSON, OutputText)
		}
		o.format = format
		return nil
	}
}

// WithTimeFormat устанавливает формат времени в логах.
// Использует стандартные форматы времени Go.
func WithTimeFormat(timeFormat string) Option {
	return func(o *options) error {
		if timeFormat == "" {
			return errors.New("time format cannot be empty")
		}
		o.timeFormat = timeFormat
		return nil
	}
}

// New создает новый экземпляр Logger с указанными опциями.
// По умолчанию используется уровень INFO, вывод в stdout, формат JSON и формат времени time.Stamp.
func New(opts ...Option) (Logger, error) {
	o := &options{
		level:      slog.LevelInfo,
		output:     os.Stdout,
		format:     OutputJSON,
		timeFormat: time.Stamp,
	}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	handlerOpts := &slog.HandlerOptions{
		Level: o.level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				return slog.String(slog.TimeKey, time.Now().Format(o.timeFormat))
			case slog.LevelKey:
				if name, ok := levelNames[a.Value.Any().(slog.Level)]; ok {
					return slog.String(slog.LevelKey, name)
				}
				if level, ok := a.Value.Any().(slog.Level); ok && level == LevelPanic {
					return slog.String(slog.LevelKey, "PNC")
				}
			}
			return a
		},
	}

	var handler slog.Handler
	switch o.format {
	case OutputJSON:
		handler = slog.NewJSONHandler(o.output, handlerOpts)
	default:
		handler = slog.NewTextHandler(o.output, handlerOpts)
	}

	return &loggerImpl{logger: slog.New(handler)}, nil
}

// With возвращает новый логгер с добавленными полями.
// Поля сортируются по ключам для обеспечения консистентного вывода.
func (l *loggerImpl) With(fields map[string]any) Logger {
	attrs := make([]any, 0, len(fields))
	for _, k := range slices.Sorted(maps.Keys(fields)) {
		attrs = append(attrs, toSlogAttr(k, fields[k]))
	}
	return &loggerImpl{logger: l.logger.With(attrs...)}
}

// Debug логирует сообщение на уровне DEBUG.
func (l *loggerImpl) Debug(msg string, args ...any) { l.logger.Debug(msg, args...) }

// Info логирует сообщение на уровне INFO.
func (l *loggerImpl) Info(msg string, args ...any) { l.logger.Info(msg, args...) }

// Warn логирует сообщение на уровне WARN.
func (l *loggerImpl) Warn(msg string, args ...any) { l.logger.Warn(msg, args...) }

// Error логирует сообщение на уровне ERROR.
func (l *loggerImpl) Error(msg string, args ...any) { l.logger.Error(msg, args...) }

// Panic логирует сообщение на уровне PANIC и завершает программу с кодом 1.
func (l *loggerImpl) Panic(msg string, args ...any) {
	l.logger.Log(context.Background(), LevelPanic, msg, args...)
	os.Exit(1)
}

// toSlogAttr преобразует ключ и значение в slog.Attr.
// Поддерживает рекурсивную обработку структур и мап, преобразуя их в группы атрибутов.
// nil-значения обрабатываются корректно.
func toSlogAttr(key string, val any) slog.Attr {
	if val == nil {
		return slog.Any(key, nil)
	}

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return slog.Any(key, nil)
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// Для time.Time возвращаем как есть, без разворачивания полей
		if v.Type() == reflect.TypeFor[time.Time]() {
			return slog.Any(key, val)
		}
		// Для остальных структур разворачиваем экспортируемые поля
		attrs := make([]slog.Attr, 0, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if f.IsExported() {
				attrs = append(attrs, toSlogAttr(f.Name, v.Field(i).Interface()))
			}
		}
		return attrsToGroup(key, attrs)

	case reflect.Map:
		// Собираем все строковые ключи мапы
		keys := make([]string, 0, v.Len())
		for _, k := range v.MapKeys() {
			if k.Kind() == reflect.String {
				keys = append(keys, k.String())
			}
		}
		slices.Sort(keys)
		attrs := make([]slog.Attr, 0, len(keys))
		for _, k := range keys {
			attrs = append(attrs, toSlogAttr(k, v.MapIndex(reflect.ValueOf(k)).Interface()))
		}
		return attrsToGroup(key, attrs)

	default:
		return slog.Any(key, val)
	}
}

// attrsToGroup создает группу атрибутов с указанным ключом.
func attrsToGroup(key string, attrs []slog.Attr) slog.Attr {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return slog.Group(key, args...)
}
