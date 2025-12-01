package configutils

import (
	"reflect"
	"strconv"
	"strings"
)

type TransformTarget struct {
	OutputKey      string
	Transform      string
	ChildTransform string
	Value          any
}

// Apply Transform funcs based on `transform`, `childtransform` and `aliases` in struct schema
func (config Config) ApplyTransformFuncs(id string, schema any, path string, funcs map[string]func(string, any) (string, any)) {
	raw := config.Layer.Get(path)

	flat := map[string]any{}
	Flatten("", raw, flat)

	targets := BuildTransformMap(id, schema)

	transformed := ApplyTransforms(flat, targets, funcs)

	result := Unflatten(transformed)

	config.Layer.Delete("")
	config.Load(result, path)
}

func ApplyTransforms(flat map[string]any, targets map[string]TransformTarget, funcs map[string]func(string, any) (string, any)) map[string]any {
	out := map[string]any{}

	for key, val := range flat {
		lower := strings.ToLower(key)

		target, ok := targets[lower]
		if !ok {
			target = findChildTransform(lower, targets)

			// fallback to default
			if target.Transform == "" {
				target.Transform = "default"
			}

			if target.OutputKey == "" {
				target.OutputKey = key
			}
		}

		newKey := key
		newValue := val

		outputKey := target.OutputKey

		fnList := strings.Split(target.Transform, ",")
		for _, fnName := range fnList {
			fnName = strings.TrimSpace(fnName)

			if fnName == "" {
				continue
			}

			fn, ok := funcs[fnName]
			if !ok {
				fn = funcs["default"]
			}

			_, last := splitPath(outputKey)
			outputKey, newValue = fn(last, newValue)

			newKey = key + DELIM + outputKey
		}

		out[newKey] = newValue
	}

	return out
}

func findChildTransform(key string, targets map[string]TransformTarget) TransformTarget {
	parts := strings.Split(key, DELIM)

	for i := len(parts) - 1; i > 0; i-- {
		parent := strings.Join(parts[:i], DELIM)

		t, ok := targets[parent];

		if ok {

			if t.ChildTransform != "" {
				return TransformTarget{
					OutputKey: parent + DELIM + parts[i],
					Transform: t.ChildTransform,
				}
			}
		}
	}

	return TransformTarget{}
}

func BuildTransformMap(id string, schema any) map[string]TransformTarget {
	out := map[string]TransformTarget{}

	getTransformMap(id, schema, "", out)

	return out
}

func getTransformMap(id string, schema any, prefix string, out map[string]TransformTarget) {
	if schema == nil {
		return
	}

	v := reflect.ValueOf(schema)
	t := reflect.TypeOf(schema)

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		base := f.Tag.Get("koanf")
		if base == "" {
			continue
		}

		aliasesRaw := getFieldWithID(id, "aliases", f.Tag)

		var aliases []string
		
		if aliasesRaw != "" {
			aliases = strings.Split(aliasesRaw, ",")
		}

		transform := getFieldWithID(id, "transform", f.Tag)
		childTransform := getFieldWithID(id, "childtransform", f.Tag)

		allKeys := append([]string{base}, aliases...)

		for _, key := range allKeys {
			if key == "" {
				continue
			}

			key = strings.ToLower(key)

			var fullKey string
			if strings.HasPrefix(key, ".") {
				fullKey = key[1:]
			} else if prefix != "" {
				fullKey = prefix + DELIM + key
			} else {
				fullKey = key
			}

			out[fullKey] = TransformTarget{
				OutputKey:      strings.ToLower(prefix + DELIM + base),
				Transform:      transform,
				ChildTransform: childTransform,
				Value:          getValueSafe(fv),
			}
		}

		// Recurse into nested structs
		if fv.Kind() == reflect.Struct || (fv.Kind() == reflect.Ptr && fv.Elem().Kind() == reflect.Struct) {

			nextPrefix := base
			if prefix != "" {
				nextPrefix = prefix + DELIM + base
			}

			getTransformMap(id, fv.Interface(), nextPrefix, out)
		}
	}
}

func Flatten(prefix string, v any, out map[string]any) {
	switch x := v.(type) {

	case map[string]any:
		for k, value := range x {
			var key string
			if prefix != "" {
				key = prefix + DELIM + k
			} else {
				key = k
			}
			Flatten(key, value, out)
		}

	case []any:
		for i, value := range x {
			key := prefix + DELIM + strconv.Itoa(i)
			Flatten(key, value, out)
		}

	default:
		out[prefix] = x
	}
}

func Unflatten(flat map[string]any) map[string]any {
	root := map[string]any{}

	for full, val := range flat {
		parts := strings.Split(full, DELIM)
		m := root

		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]

			_, ok := m[part]; 

			if !ok {
				m[part] = map[string]any{}
			}

			m = m[part].(map[string]any)
		}

		m[parts[len(parts)-1]] = val
	}

	return root
}

func splitPath(p string) (string, string) {
	parts := strings.Split(p, DELIM)

	if len(parts) == 1 {
		return "", p
	}

	return strings.Join(parts[:len(parts)-1], DELIM), parts[len(parts)-1]
}

func getValueSafe(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return getValueSafe(v.Elem())
	}
	return v.Interface()
}

func getFieldWithID(id string, key string, tag reflect.StructTag) string {
	if id != "" {
		value, ok := tag.Lookup(id + ">" + key);

		if ok {
			return value
		}
	}

	return tag.Get(key)
}