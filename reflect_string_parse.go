package redisobj

import (
	"fmt"
	"reflect"
	"strconv"
)

func setFieldFromString(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.Set(reflect.ValueOf(value))
		return nil
	case reflect.Bool:
		if value == "" {
			val := reflect.ValueOf(false)
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseBool(value); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Int:
		if value == "" {
			val := reflect.ValueOf(int(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			val := reflect.ValueOf(int(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Int8:
		if value == "" {
			val := reflect.ValueOf(int8(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseInt(value, 10, 8); err == nil {
			val := reflect.ValueOf(int8(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Int16:
		if value == "" {
			val := reflect.ValueOf(int16(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseInt(value, 10, 16); err == nil {
			val := reflect.ValueOf(int16(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Int32:
		if value == "" {
			val := reflect.ValueOf(int32(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			val := reflect.ValueOf(int32(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Int64:
		if value == "" {
			val := reflect.ValueOf(int64(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			val := reflect.ValueOf(int64(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Uint:
		if value == "" {
			val := reflect.ValueOf(uint(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			val := reflect.ValueOf(uint(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Uint8:
		if value == "" {
			val := reflect.ValueOf(uint8(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseUint(value, 10, 8); err == nil {
			val := reflect.ValueOf(uint8(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Uint16:
		if value == "" {
			val := reflect.ValueOf(uint16(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseUint(value, 10, 16); err == nil {
			val := reflect.ValueOf(uint16(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Uint32:
		if value == "" {
			val := reflect.ValueOf(uint32(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseUint(value, 10, 32); err == nil {
			val := reflect.ValueOf(uint32(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Uint64:
		if value == "" {
			val := reflect.ValueOf(uint64(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			val := reflect.ValueOf(uint64(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Float32:
		if value == "" {
			val := reflect.ValueOf(float32(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseFloat(value, 32); err == nil {
			val := reflect.ValueOf(float32(parsedValue))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	case reflect.Float64:
		if value == "" {
			val := reflect.ValueOf(float64(0))
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		} else if parsedValue, err := strconv.ParseFloat(value, 64); err == nil {
			val := reflect.ValueOf(parsedValue)
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return nil
			}
		}
	}

	return fmt.Errorf("%w: could not set value (%s) from string (%s)", ErrInvalidFieldType, field.Kind(), value)
}

func valueToString(value reflect.Value) (string, error) {
	switch typedValue := value.Interface().(type) {
	case string:
		return typedValue, nil
	case bool:
		return strconv.FormatBool(typedValue), nil
	case int:
		return strconv.FormatInt(int64(typedValue), 10), nil
	case int8:
		return strconv.FormatInt(int64(typedValue), 10), nil
	case int16:
		return strconv.FormatInt(int64(typedValue), 10), nil
	case int32:
		return strconv.FormatInt(int64(typedValue), 10), nil
	case int64:
		return strconv.FormatInt(int64(typedValue), 10), nil
	case uint:
		return strconv.FormatUint(uint64(typedValue), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(typedValue), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(typedValue), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(typedValue), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(typedValue), 10), nil
	case float32:
		return strconv.FormatFloat(float64(typedValue), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(float64(typedValue), 'f', -1, 64), nil
	}

	return "", fmt.Errorf("%w: could not convert value to string: %v", ErrInvalidFieldType, value.Interface())
}

func isStringParsable(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String:
		return true
	case reflect.Bool:
		return true
	case reflect.Int:
		return true
	case reflect.Int8:
		return true
	case reflect.Int16:
		return true
	case reflect.Int32:
		return true
	case reflect.Int64:
		return true
	case reflect.Uint:
		return true
	case reflect.Uint8:
		return true
	case reflect.Uint16:
		return true
	case reflect.Uint32:
		return true
	case reflect.Uint64:
		return true
	case reflect.Float32:
		return true
	case reflect.Float64:
		return true
	}

	return false
}
