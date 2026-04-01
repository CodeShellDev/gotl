package jsonutils

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
)

// Get data by supplying a path
func GetByPath(path string, data any) (any, bool) {
	// Split into parts by `.` and `[]`
	re := regexp.MustCompile(`\.|\[|\]`)

	parts := re.Split(path, -1)

	cleaned := []string{}

	for _, part := range parts {
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}

	current := data

	for _, key := range cleaned {
		switch currentDataType := current.(type) {
		case map[string]any:
			value, ok := currentDataType[key]
			if !ok {
				return nil, false
			}
			current = value

		case []any:
			index, err := strconv.Atoi(key)

			if err != nil || index < 0 || index >= len(currentDataType) {
				return nil, false
			}
			current = currentDataType[index]

		default:
			return nil, false
		}
	}

	return current, true
}

// Get data from json string (returns error)
func GetJsonSafe[T any](str string) (T, error) {
	var result T

	err := json.Unmarshal([]byte(str), &result)

	return result, err
}

// Get data from json string (without error)
func GetJson[T any](str string) T {
	var result T

	json.Unmarshal([]byte(str), &result)

	return result
}

// Get json string from data (returns error)
func ToJsonSafe(obj any) (string, error) {
	bytes, err := json.Marshal(obj)

	return string(bytes), err
}

// Get json string from data (without error)
func ToJson(obj any) string {
	bytes, _ := json.Marshal(obj)

	return string(bytes)
}

// Get json string from data (without error and ignoring unmarshable fields)
func ToJsonSkipIncompatible(obj any) string {
	bytes, _ := MarshalSkipIncompatible(obj)

	return string(bytes)
}

// Prettify data into json string
func Pretty[T any](obj T) string {
	bytes, _ := json.MarshalIndent(obj, "", "  ")

	return string(bytes)
}

// Marshals an object into json bytes, while ignoring unmarshable fields according to json.Marshal()
func MarshalSkipIncompatible(obj any) ([]byte, error) {
	clean := sanitize(reflect.ValueOf(obj))
	return json.Marshal(clean)
}

func sanitize(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}

	// unwrap pointers and interfaces
	for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}

		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		out := map[string]any{}

		t := value.Type()

		for i := 0; i < value.NumField(); i++ {
			field := t.Field(i)

			// skip unexported
			if field.PkgPath != "" {
				continue
			}

			name := field.Name

			fieldValue := value.Field(i)
			cleaned := sanitize(fieldValue)

			if cleaned != nil {
				out[name] = cleaned
			}
		}

		return out

	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return nil // json only supports string keys
		}

		out := map[string]any{}

		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			iterValue := iter.Value()

			cleaned := sanitize(iterValue)
			if cleaned != nil {
				out[key] = cleaned
			}
		}

		return out

	case reflect.Slice, reflect.Array:
		out := make([]any, 0, value.Len())

		for i := 0; i < value.Len(); i++ {
			cleaned := sanitize(value.Index(i))
			if cleaned != nil {
				out = append(out, cleaned)
			}
		}

		return out

	default:
		val := value.Interface()

		_, err := json.Marshal(val)
		if err == nil {
			return val
		}
		
		return nil
	}
}