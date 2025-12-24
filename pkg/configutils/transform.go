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

	for key, value := range flat {
		parts := splitPath(key)

		var current any = root
		var parent any
		var parentKey any

		for i, part := range parts {
			last := i == len(parts) - 1

			intPart, err := strconv.Atoi(part)

			if err == nil {
				var slice []any

				if current == nil {
					slice = []any{}
				} else {
					slice = current.([]any)
				}

				for len(slice) <= intPart {
					slice = append(slice, nil)
				}

				// save slice in parent
				switch asserted := parent.(type) {
				case map[string]any:
					asserted[parentKey.(string)] = slice
				case []any:
					asserted[parentKey.(int)] = slice
				}

				if last {
					slice[intPart] = value
					break
				}

				if slice[intPart] == nil {
					_, err := strconv.Atoi(parts[i+1])
					if err == nil {
						slice[intPart] = []any{}
					} else {
						slice[intPart] = map[string]any{}
					}
				}

				parent = slice
				parentKey = intPart
				current = slice[intPart]
				continue
			} else {
				m, ok := current.(map[string]any)
				if !ok {
					m = map[string]any{}
				}

				if last {
					m[part] = value
					break
				}

				_, exists := m[part]
				if !exists {
					_, err := strconv.Atoi(parts[i+1])
					if err == nil {
						m[part] = []any{}
					} else {
						m[part] = map[string]any{}
					}
				}

				parent = m
				parentKey = part
				current = m[part]
			}
		}
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

	result := Unflatten(transformed)

	config.Layer.Delete("")
	config.Load(result, path)
}

func ApplyTransforms(flat map[string]any, targets map[string]TransformTarget, funcs map[string]func(string, any) (string, any)) map[string]any {
	out := map[string]any{}

	for key, val := range flat {
		keyParts := splitPath(key)

		newKeyParts := []string{}
		newValue := val

		fullTarget := resolveTransform(strings.ToLower(key), targets)

		fmt.Println("Target: ", jsonutils.Pretty(fullTarget))

		if fullTarget.OutputKey != "" && len(keyParts) != len(splitPath(fullTarget.OutputKey)) {
			key = fullTarget.OutputKey

			keyParts = splitPath(key)
		}

		for i := range keyParts {
			parent := joinPaths(keyParts[:i+1]...)
			lower := strings.ToLower(parent)

			target := resolveTransform(lower, targets)

			// fallback to default
			if target.Transform == "" {
				target.Transform = "default"
			}
			if target.OutputKey == "" {
				target.OutputKey = parent
			}

			fmt.Println("NewTarget: ", jsonutils.Pretty(target))

			outputKeyParts := splitPath(target.OutputKey)
			outputBase := outputKeyParts[len(outputKeyParts)-1]

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

				if fn == nil {
					continue
				}

				outputBase, newValue = fn(outputBase, newValue)
			}

			newKeyParts = append(newKeyParts, outputBase)
		}

		fmt.Println(key, " => ", joinPaths(newKeyParts...))

		out[joinPaths(newKeyParts...)] = newValue
	}

	return out
}

func resolveTransform(lower string, targets map[string]TransformTarget) TransformTarget {
	t, ok := targets[lower]

	if ok {
        return t
    }

	t = findTransform(lower, targets)

    if t.Transform != "" {
        return TransformTarget{
            OutputKey:      t.OutputKey,
            Transform:      t.Transform,
            ChildTransform: t.ChildTransform,
        }
    }

    parts := splitPath(lower)
    for i := len(parts) - 1; i >= 1; i-- {
        parent := joinPaths(parts[:i]...)

        t := findTransform(parent, targets)
        if t.ChildTransform != "" {
            fullKey := joinPaths(t.OutputKey, joinPaths(parts[i:]...))

            return TransformTarget{
                OutputKey:      fullKey,
                Transform:      t.ChildTransform,
                ChildTransform: t.ChildTransform,
            }
        }
    }

    return TransformTarget{}
}

func findTransform(lower string, targets map[string]TransformTarget) TransformTarget {
    actualParts := splitPath(lower)

    bestLen := -1
    var best TransformTarget

    for schema, t := range targets {
        schemaParts := splitPath(schema)

        if matchWithDynamic(actualParts, schemaParts) {
            if len(schemaParts) > bestLen {
                bestLen = len(schemaParts)
                best = t
            }
        }
    }

    return best
}

func matchWithDynamic(actualParts, schemaParts []string) bool {
    if len(actualParts) < len(schemaParts) {
        return false
    }

    offset := len(actualParts) - len(schemaParts)

    for i := range schemaParts {
        schemaPart := schemaParts[i]
        actualPart := actualParts[i+offset]

        if schemaPart == "*" {
            continue
        }

        if !strings.EqualFold(schemaPart, actualPart) {
            return false
        }
    }

    return true
}

func splitPath(p string) []string {
	return strings.Split(p, DELIM)
}

func joinPaths(p ...string) string {
	return strings.Join(p, DELIM)
}