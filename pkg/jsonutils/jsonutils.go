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

// Get data from jsonStr (returns error)
func GetJsonSafe[T any](jsonStr string) (T, error) {
	var result T

	err := json.Unmarshal([]byte(jsonStr), &result)

	return result, err
}

// Get data from jsonStr (without error)
func GetJson[T any](jsonStr string) T {
	var result T

	json.Unmarshal([]byte(jsonStr), &result)

	return result
}

// Get jsonStr from data (returns error)
func ToJsonSafe(obj any) (string, error) {
	bytes, err := json.Marshal(obj)

	return string(bytes), err
}

// Get jsonStr from data (without error)
func ToJson(obj any) string {
	bytes, _ := json.Marshal(obj)

	return string(bytes)
}

// Prettify data into jsonStr
func Pretty[T any](obj T) string {
	bytes, _ := json.MarshalIndent(obj, "", "  ")

	return string(bytes)
}

func PrettySkipIncompatible(obj any) string {
	bytes, err := MarshalSkipIncompatible(obj)

	if err != nil {
		return ""
	}

	cleaned := GetJson[any](string(bytes))

	bytes, err = json.MarshalIndent(cleaned, "", "  ")

	if err != nil {
		return ""
	}

	return string(bytes)
}

// Marshals an object into json bytes, while ignoring unmarshable fields according to json.Marshal()
func MarshalSkipIncompatible(v any) ([]byte, error) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// handle pointer
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return []byte("null"), nil
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return json.Marshal(v)
	}

	out := make(map[string]any)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		name := field.Name

		fv := val.Field(i).Interface()

		// try marshaling field

		_, err := json.Marshal(fv)
		if err == nil {
			out[name] = fv
		}
	}

	return json.Marshal(out)
}