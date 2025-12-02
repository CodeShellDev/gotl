package configutils

import (
	"reflect"
	"strings"
)

// Build transform map
func BuildTransformMap(id string, schema any) map[string]TransformTarget {
	out := map[string]TransformTarget{}

	getTransformMap(id, schema, "", out)

	return out
}

func getTransformMap(id string, schema any, stem string, out map[string]TransformTarget) {
	if schema == nil {
		return
	}

	v := reflect.ValueOf(schema)
	t := reflect.TypeOf(schema)

	if t.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}

		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		base := field.Tag.Get("koanf")
		if base == "" {
			continue
		}

		aliasesRaw := getFieldWithID(id, "aliases", field.Tag)

		var aliases []string
		if aliasesRaw != "" {
			aliases = strings.Split(aliasesRaw, ",")
		}

		transform := getFieldWithID(id, "transform", field.Tag)
		childTransform := getFieldWithID(id, "childtransform", field.Tag)

		allKeys := append([]string{base}, aliases...)

		for _, key := range allKeys {
			if key == "" {
				continue
			}

			key = strings.ToLower(key)

			var fullKey string
			outputKey := stem + DELIM + base

			if strings.HasPrefix(key, ".") {
				fullKey = key[1:]
			} else if stem != "" {
				fullKey = stem + DELIM + key
			} else {
				fullKey = key
				outputKey = base
			}

			out[fullKey] = TransformTarget{
				OutputKey:      strings.ToLower(outputKey),
				Transform:      transform,
				ChildTransform: childTransform,
				Value:          getValueSafe(fieldValue),
			}
		}

		nextStem := base
		if stem != "" {
			nextStem = stem + DELIM + base
		}

		fieldKind := fieldValue.Kind()

		switch fieldKind {
		case reflect.Struct:
			getTransformMap(id, fieldValue.Interface(), nextStem, out)
		case reflect.Pointer:
			handlePointer(id, fieldValue, nextStem, out)
		case reflect.Slice, reflect.Array:
			handleArray(id, field, nextStem, out)
		case reflect.Map:
			handleMap(id, field, nextStem, out)
		}
	}
}

func handlePointer(id string, fieldValue reflect.Value, stem string, out map[string]TransformTarget) {
	if !fieldValue.IsNil() {
		elem := fieldValue.Elem()

		if elem.Kind() == reflect.Struct {
			getTransformMap(id, elem.Interface(), stem, out)
		}
	}
}

func handleArray(id string, field reflect.StructField, stem string, out map[string]TransformTarget) {
	t := field.Type.Elem()
	k := t.Kind()

	if k == reflect.Pointer {
		t = t.Elem()
		k = t.Kind()
	}

	if k == reflect.Struct {
		zero := reflect.New(t).Elem().Interface()
		
		getTransformMap(id, zero, stem, out)
	}
}

func handleMap(id string, field reflect.StructField, stem string, out map[string]TransformTarget) {
	t := field.Type.Key().Kind()

	if t == reflect.String {
		valueType := field.Type.Elem()
		valueKind := valueType.Kind()

		if valueKind == reflect.Pointer {
			valueType = valueType.Elem()
			valueKind = valueType.Kind()
		}

		if valueKind == reflect.Struct {
			iface := reflect.New(valueType).Elem().Interface()

			getTransformMap(id, iface, stem, out)
		}
	}
}