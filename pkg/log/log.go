package log

import (
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	OutputJSON = "json"
	OutputText = "text"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
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
	addSource  bool
}

type Option func(*options) error

func WithLevel(level string) Option {
	m := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	return func(o *options) error {
		o.level = m[strings.ToLower(level)]
		return nil
	}
}

func WithOutput(w io.Writer) Option {
	return func(o *options) error {
		o.output = w
		return nil
	}
}

func WithFormat(format string) Option {
	return func(o *options) error {
		if format != OutputJSON && format != OutputText {
			format = OutputJSON
		}
		o.format = format
		return nil
	}
}

func WithTimeFormat(timeFormat string) Option {
	return func(o *options) error {
		o.timeFormat = timeFormat
		return nil
	}
}

func WithAddSource(add bool) Option {
	return func(o *options) error {
		o.addSource = add
		return nil
	}
}

func New(opts ...Option) (Logger, error) {
	o := &options{
		level:      slog.LevelInfo,
		output:     os.Stdout,
		format:     OutputJSON,
		timeFormat: time.Stamp,
		addSource:  false,
	}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	optsHandler := &slog.HandlerOptions{
		Level:     o.level,
		AddSource: o.addSource,
	}

	optsHandler.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey:
			return slog.String(slog.TimeKey, time.Now().Format(o.timeFormat))
		case slog.LevelKey:
			short := map[slog.Level]string{
				slog.LevelDebug: "DBG",
				slog.LevelInfo:  "INF",
				slog.LevelWarn:  "WRN",
				slog.LevelError: "ERR",
			}
			if v, ok := short[a.Value.Any().(slog.Level)]; ok {
				return slog.String(slog.LevelKey, v)
			}
		}
		return a
	}

	var handler slog.Handler
	switch o.format {
	case OutputJSON:
		handler = slog.NewJSONHandler(o.output, optsHandler)
	default:
		handler = slog.NewTextHandler(o.output, optsHandler)
	}

	return &loggerImpl{
		logger: slog.New(handler),
	}, nil
}

func (l *loggerImpl) With(fields map[string]any) Logger {
	attrs := make([]any, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, toSlogAttr(k, v))
	}
	return &loggerImpl{
		logger: l.logger.With(attrs...),
	}
}

// toSlogAttr рекурсивно преобразует значение в slog.Attr,
// создавая группы для структур и карт
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
		if v.Type() == reflect.TypeFor[time.Time]() {
			return slog.Any(key, val)
		}

		attrs := make([]slog.Attr, 0, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			attrs = append(attrs, toSlogAttr(field.Name, v.Field(i).Interface()))
		}

		args := make([]any, len(attrs))
		for i, attr := range attrs {
			args[i] = attr
		}
		return slog.Group(key, args...)

	case reflect.Map:
		attrs := make([]slog.Attr, 0, v.Len())
		for _, k := range v.MapKeys() {
			if k.Kind() != reflect.String {
				continue
			}
			attrs = append(attrs, toSlogAttr(k.String(), v.MapIndex(k).Interface()))
		}

		args := make([]any, len(attrs))
		for i, attr := range attrs {
			args[i] = attr
		}
		return slog.Group(key, args...)

	default:
		return slog.Any(key, val)
	}
}

func (l *loggerImpl) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *loggerImpl) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *loggerImpl) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *loggerImpl) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}
