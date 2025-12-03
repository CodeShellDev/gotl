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
				key = joinPaths(prefix, k)
			} else {
				key = k
			}

			Flatten(key, value, out)
		}

	case []any:
		for i, value := range asserted {
			key := joinPaths(prefix, strconv.Itoa(i))

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
		parts := splitPath(full)
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

	targets := BuildTransformMap(id, schema)

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
		keyParts := splitPath(key)

		newKeyParts := []string{}

		newValue := val

		for i := range keyParts {
			parent := joinPaths(keyParts[:i]...)

			if i == len(keyParts) - 1 {
				parent = key
			}

			fmt.Println(i, key, keyParts, parent)
			
			lower := strings.ToLower(parent)

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

			outputKey := target.OutputKey

			outputkeyParts := splitPath(outputKey)

			outputBase := outputkeyParts[len(outputkeyParts)-1]

			fnList := strings.SplitSeq(target.Transform, ",")
			for fnName := range fnList {
				fnName = strings.TrimSpace(fnName)

				if fnName == "" {
					continue
				}

				fn, ok := funcs[fnName]
				if !ok {
					fn = funcs["default"]
				}

				outputBase, newValue = fn(outputBase, newValue)
			}

			fmt.Println(outputKey, outputBase)

			newKeyParts = append(newKeyParts, outputBase)
		}

		out[joinPaths(newKeyParts...)] = newValue
	}

	return out
}

func findChildTransform(key string, targets map[string]TransformTarget) TransformTarget {
	parts := splitPath(key)

	for i := len(parts) - 1; i > 0; i-- {
		parent := joinPaths(parts[:i]...)

		t, ok := targets[parent]

		if ok {
			if t.ChildTransform != "" {
				return TransformTarget{
					OutputKey: joinPaths(parent, parts[i]),
					Transform: t.ChildTransform,
				}
			}
		}
	}

	return TransformTarget{}
}

func splitPath(p string) []string {
	return strings.Split(p, DELIM)
}

func joinPaths(p ...string) string {
	return strings.Join(p, DELIM)
}