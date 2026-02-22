package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type field struct {
	index      []int
	envKey     string
	required   bool
	defValue   string
	sep        string
	isDuration bool
	typ        reflect.Type
}

// Load загружает конфигурацию из .env файлов и переменных окружения
func Load(cfg any, files ...string) error {
	envMap, err := buildEnvMap(files, true)
	if err != nil {
		return err
	}
	return apply(cfg, envMap)
}

// LoadEnv загружает конфигурацию только из переменных окружения
func LoadEnv(cfg any) error {
	envMap, err := buildEnvMap(nil, true)
	if err != nil {
		return err
	}
	return apply(cfg, envMap)
}

// LoadFile загружает конфигурацию только из .env файлов
func LoadFile(cfg any, files ...string) error {
	envMap, err := buildEnvMap(files, false)
	if err != nil {
		return err
	}
	return apply(cfg, envMap)
}

// buildEnvMap создает карту переменных из файлов и/или системы
func buildEnvMap(files []string, includeSystem bool) (map[string]string, error) {
	envMap := make(map[string]string)

	for _, file := range files {
		if err := loadFile(envMap, file); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("load file %s: %w", file, err)
			}
		}
	}

	if includeSystem {
		for _, kv := range os.Environ() {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	return envMap, nil
}

// loadFile читает .env файл и добавляет пары ключ-значение в карту
func loadFile(dest map[string]string, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(value) > 1 && (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		dest[key] = value
	}

	return nil
}

// apply применяет значения из карты к структуре конфигурации
func apply(cfg any, envMap map[string]string) error {
	val := reflect.ValueOf(cfg)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("cfg must be a non-nil pointer to struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("cfg must point to a struct")
	}

	typ := val.Type()
	infos, err := getfileds(typ)
	if err != nil {
		return err
	}

	var errs []error
	for _, info := range infos {
		field := val.FieldByIndex(info.index)

		rawValue, ok := envMap[info.envKey]
		if !ok || rawValue == "" {
			if info.defValue != "" {
				rawValue = info.defValue
			} else if info.required {
				errs = append(errs, fmt.Errorf("missing required env: %s", info.envKey))
				continue
			} else {
				continue
			}
		}

		if err := setFieldValue(field, rawValue, info); err != nil {
			errs = append(errs, fmt.Errorf("field %s: %w", info.envKey, err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// getfileds возвращает метаданные полей структуры
func getfileds(typ reflect.Type) ([]field, error) {
	infos, err := buildfileds(typ, nil, "")
	if err != nil {
		return nil, err
	}

	return infos, nil
}

// buildfileds рекурсивно собирает информацию о полях структуры
func buildfileds(typ reflect.Type, path []int, prefix string) ([]field, error) {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", typ.Kind())
	}

	var infos []field
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("env")
		if tag == "-" {
			continue
		}

		key, required, defValue, sep := parseEnvTag(tag)

		currentPath := append(path, i)
		envKey := buildEnvKey(prefix, f.Name, key)

		if f.Type.Kind() == reflect.Struct && f.Type != reflect.TypeFor[time.Time]() {
			nested, err := buildfileds(f.Type, currentPath, envKey+"_")
			if err != nil {
				return nil, err
			}
			infos = append(infos, nested...)
			continue
		}
		if f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct && f.Type.Elem() != reflect.TypeFor[time.Time]() {
			nested, err := buildfileds(f.Type.Elem(), currentPath, envKey+"_")
			if err != nil {
				return nil, err
			}
			infos = append(infos, nested...)
			continue
		}

		info := field{
			index:      currentPath,
			envKey:     envKey,
			required:   required,
			defValue:   defValue,
			sep:        sep,
			isDuration: f.Type == reflect.TypeFor[time.Duration](),
			typ:        f.Type,
		}

		if info.sep == "" {
			info.sep = ","
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// parseEnvTag разбирает тег env на составные части
func parseEnvTag(tag string) (key string, required bool, defValue string, sep string) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return "", false, "", ","
	}

	if parts[0] != "" {
		key = parts[0]
	}

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		switch {
		case part == "required":
			required = true

		case strings.HasPrefix(part, "default="):
			defValue = strings.TrimPrefix(part, "default=")
			defValue = strings.Trim(defValue, "\"'")

		case strings.HasPrefix(part, "sep="):
			sep = strings.TrimPrefix(part, "sep=")
			sep = strings.Trim(sep, "\"'")
		}
	}

	if sep == "" {
		sep = ","
	}

	return
}

// buildEnvKey формирует ключ переменной окружения
func buildEnvKey(prefix, fieldName, key string) string {
	if key != "" {
		return strings.ToUpper(prefix + key)
	}
	return strings.ToUpper(prefix + strings.ToUpper(fieldName))
}

// setFieldValue устанавливает значение поля структуры
func setFieldValue(field reflect.Value, rawValue string, info field) error {
	if !field.CanSet() {
		return nil
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	if info.isDuration {
		d, err := time.ParseDuration(rawValue)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}

		field.Set(reflect.ValueOf(d))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(rawValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int: %w", err)
		}
		field.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(rawValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint: %w", err)
		}
		field.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %w", err)
		}
		field.SetFloat(v)

	case reflect.Bool:
		v, err := strconv.ParseBool(rawValue)
		if err != nil {
			return fmt.Errorf("invalid bool: %w", err)
		}
		field.SetBool(v)

	case reflect.Slice:
		return setSliceValue(field, rawValue, info.sep)

	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}

	return nil
}

// setSliceValue заполняет срез значениями из строки с разделителем
func setSliceValue(field reflect.Value, rawValue string, sep string) error {
	items := strings.Split(rawValue, sep)
	slice := reflect.MakeSlice(field.Type(), len(items), len(items))

	for i, item := range items {
		item = strings.TrimSpace(item)
		elem := slice.Index(i)

		switch elem.Kind() {
		case reflect.String:
			elem.SetString(item)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, err := strconv.ParseInt(item, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int in slice: %w", err)
			}
			elem.SetInt(v)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v, err := strconv.ParseUint(item, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid uint in slice: %w", err)
			}
			elem.SetUint(v)

		case reflect.Float32, reflect.Float64:
			v, err := strconv.ParseFloat(item, 64)
			if err != nil {
				return fmt.Errorf("invalid float in slice: %w", err)
			}
			elem.SetFloat(v)

		case reflect.Bool:
			v, err := strconv.ParseBool(item)
			if err != nil {
				return fmt.Errorf("invalid bool in slice: %w", err)
			}
			elem.SetBool(v)

		default:
			return fmt.Errorf("unsupported slice element type %s", elem.Kind())
		}
	}

	field.Set(slice)
	return nil
}
