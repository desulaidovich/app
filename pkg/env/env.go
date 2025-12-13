package env

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Parser[T any] struct {
	config T
}

func New[T any](cfg T) *Parser[T] {
	return &Parser[T]{
		config: cfg,
	}
}

func (p *Parser[T]) Parse() (T, error) {
	return p.config, nil
}

func LoadFromEnv[T any](cfg T) (T, error) {
	return New(cfg).FromENV().Parse()
}

func (p *Parser[T]) loadStruct(val reflect.Value, prefix string) {
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		tag := fieldType.Tag.Get("env")
		if tag == "-" {
			continue
		}

		if field.Kind() == reflect.Struct && field.Type() != reflect.TypeFor[time.Time]() {
			var nestedPrefix string
			if tag != "" {
				nestedPrefix = prefix + tag + "_"
			} else if prefix == "" {
				nestedPrefix = strings.ToUpper(fieldType.Name) + "_"
			} else {
				nestedPrefix = prefix + strings.ToUpper(fieldType.Name) + "_"
			}

			p.loadStruct(field, nestedPrefix)

			continue
		}

		var envKey string
		if tag != "" {
			envKey = prefix + tag
		} else {
			envKey = prefix + strings.ToUpper(fieldType.Name)
		}

		envKey = strings.ToUpper(envKey)

		envValue := os.Getenv(envKey)

		if envValue == "" {
			continue
		}

		p.setFieldValue(field, envValue)
	}
}

func (p *Parser[T]) FromENV() *Parser[T] {
	val := reflect.ValueOf(&p.config).Elem()
	p.loadStruct(val, "")

	return p
}

func (p *Parser[T]) setFieldValue(field reflect.Value, value string) {
	if !field.CanSet() {
		return
	}

	p.settimeDurationValue(field, value)

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intVal)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintVal)
		}

	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatVal)
		}

	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		}

	case reflect.Slice:
		p.setSliceValue(field, value)
	}
}

func (p *Parser[T]) settimeDurationValue(field reflect.Value, value string) {
	if field.Type() == reflect.TypeFor[time.Duration]() {
		if d, err := time.ParseDuration(value); err == nil {
			field.Set(reflect.ValueOf(d))
		}
	}
}

func (p *Parser[T]) setSliceValue(field reflect.Value, value string) {
	items := strings.Split(value, ",")
	sliceType := field.Type()
	slice := reflect.MakeSlice(sliceType, len(items), len(items))

	for i, item := range items {
		item = strings.TrimSpace(item)
		elem := slice.Index(i)

		switch elem.Kind() {
		case reflect.String:
			elem.SetString(item)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(item, 10, 64); err == nil {
				elem.SetInt(intVal)
			}

		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(item); err == nil {
				elem.SetBool(boolVal)
			}
		}
	}

	field.Set(slice)
}
