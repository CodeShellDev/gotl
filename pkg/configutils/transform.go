package configutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/jsonutils"
)

type TransformTarget struct {
	OutputKey      string
	Transform      string
	ChildTransform string
	Value          any
}

// Flatten `data: { key: value }` into `data.key: value``
func Flatten(prefix string, v any, out map[string]any) {
	switch asserted := v.(type) {

	case map[string]any:
		for k, value := range asserted {
			var key string
			if prefix != "" {
				key = prefix + DELIM + k
			} else {
				key = k
			}

			Flatten(key, value, out)
		}

	case []any:
		for i, value := range asserted {
			key := prefix + DELIM + strconv.Itoa(i)

			Flatten(key, value, out)
		}

	default:
		out[prefix] = asserted
	}
}

// Unflatten `data.key: value` into `data: { key: value }`
func Unflatten(flat map[string]any) map[string]any {
	root := map[string]any{}

	for full, val := range flat {
		parts := strings.Split(full, DELIM)
		m := root

		for i := 0; i < len(parts) - 1; i++ {
			part := parts[i]

			_, ok := m[part] 

			if !ok {
				m[part] = map[string]any{}
			}

			m = m[part].(map[string]any)
		}

		m[parts[len(parts)-1]] = val
	}

	return root
}

// Apply Transform funcs based on `transform`, `childtransform` and `aliases` in struct schema
func (config Config) ApplyTransformFuncs(id string, schema any, path string, funcs map[string]func(string, any) (string, any)) {
	raw := config.Layer.Get(path)

	flat := map[string]any{}
	Flatten("", raw, flat)

	fmt.Println("Flattened: ", jsonutils.Pretty(flat))

	targets := BuildTransformMap(id, schema)

	fmt.Println("Targets: ", jsonutils.Pretty(targets))

	transformed := ApplyTransforms(flat, targets, funcs)

	fmt.Println("Transformed: ", jsonutils.Pretty(transformed))

	result := Unflatten(transformed)

	fmt.Println("Unflattened: ", jsonutils.Pretty(result))

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

		t, ok := targets[parent]

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

func splitPath(p string) (string, string) {
	parts := strings.Split(p, DELIM)

	if len(parts) == 1 {
		return "", p
	}

	return strings.Join(parts[:len(parts)-1], DELIM), parts[len(parts)-1]
}