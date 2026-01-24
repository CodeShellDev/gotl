package configutils

import (
	"reflect"
	"strconv"
	"strings"
)

type TransformTarget struct {
	OutputKey      	string
	Parent			string
	Source			reflect.StructField
	OnUse			string
	Transform     	string
	ChildTransform	string
	Value          	any
}

type TransformOptions struct {
	Transforms		map[string]func(key string, value any) (string, any)
	OnUse			map[string]func(source string, target TransformTarget)
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
func (config Config) ApplyTransformFuncs(id string, schema any, path string, options TransformOptions) {
	raw := config.Layer.Get(path)

	flat := map[string]any{}
	Flatten("", raw, flat)

	targets := BuildTransformMap(id, schema)

	transformed := ApplyTransforms(flat, targets, options)

	result := Unflatten(transformed)

	config.Layer.Delete("")
	config.Load(result, path)
}

func ApplyTransforms(flat map[string]any, targets map[string]TransformTarget, options TransformOptions) map[string]any {
	out := map[string]any{}

	for key, val := range flat {
		keyParts := splitPath(key)

		newKeyParts := []string{}
		newValue := val

		source, fullTarget := resolveTransform(strings.ToLower(key), targets)

		if fullTarget.OutputKey != "" && len(keyParts) != len(splitPath(fullTarget.OutputKey)) {
			key = fullTarget.OutputKey

			keyParts = splitPath(key)
		}

		var target TransformTarget

		for i := range keyParts {
			parent := joinPaths(keyParts[:i+1]...)
			lower := strings.ToLower(parent)

			var match string
			match, target = resolveTransform(lower, targets)

			if source != "" {
				match = source
			}

			// fallback to default
			if target.Transform == "" {
				target.Transform = "default"
			}
			if target.OutputKey == "" {
				target.OutputKey = parent
			}

			outputKeyParts := splitPath(target.OutputKey)
			outputBase := outputKeyParts[len(outputKeyParts)-1]

			transformList := strings.SplitSeq(target.Transform, ",")
			for fnName := range transformList {
				fnName = strings.TrimSpace(fnName)

				if fnName == "" {
					continue
				}

				fn, ok := options.Transforms[fnName]
				if !ok {
					fn = options.Transforms["default"]
				}

				if fn == nil {
					continue
				}

				outputBase, newValue = fn(outputBase, newValue)
			}

			// fallback to default
			if target.OnUse == "" {
				target.OnUse = "default"
			}

			onUseMap := ParseTag(target.OnUse)

			onUse := GetValueWithSource(match, target.Parent, onUseMap)

			onUseList := strings.SplitSeq(onUse, ",")
			for fnName := range onUseList {
				fnName = strings.TrimSpace(fnName)

				if fnName == "" {
					continue
				}

				fn, ok := options.OnUse[fnName]
				if !ok {
					continue
				}

				if fn == nil {
					continue
				}

				fn(match, target)
			}

			newKeyParts = append(newKeyParts, outputBase)
		}

		out[joinPaths(newKeyParts...)] = newValue
	}

	return out
}

func resolveTransform(lower string, targets map[string]TransformTarget) (string, TransformTarget) {
	t, ok := targets[lower]

	if ok {
        return lower, t
    }

	t = findTransform(lower, targets)

    if t.Transform != "" {
        return lower, TransformTarget{
            OutputKey:      t.OutputKey,
			Parent: 		t.Parent,
			Source: 		t.Source,
            Transform:      t.Transform,
			OnUse: 			t.OnUse,
            ChildTransform: t.ChildTransform,
        }
    }

    parts := splitPath(lower)
    for i := len(parts) - 1; i >= 1; i-- {
        parent := joinPaths(parts[:i]...)

        t := findTransform(parent, targets)
		
		if isContainer(t.Value) {
            fullKey := joinPaths(t.OutputKey, joinPaths(parts[i:]...))

            return parent, TransformTarget{
                OutputKey:      fullKey,
				Parent: 		t.Parent,	
				Source: 		t.Source,
                Transform:      t.ChildTransform,
				OnUse: 			t.OnUse,
                ChildTransform: t.ChildTransform,
            }
        }
    }

    return "", TransformTarget{}
}

func isContainer(v any) bool {
	if v == nil {
		return false
	}

	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
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

func GetValueWithSource(source, parent string, valueMap map[string]string) string {
	if !strings.HasPrefix(source, parent) {
		parent = ""
	}

	if parent == "" {
		value, exists := valueMap["." + source]

		if exists {
			return value
		}
	}

	base, ok := strings.CutPrefix(source, parent + ".")

	if ok {
		value, exists := valueMap[base]

		if exists {
			return value
		}
	}

	return valueMap["*"]
}

func parseTagPart(part string) ([]string, string) {
	s := []string{}
	
	str, value, exists := strings.Cut(part, ">>")

	searchList := strings.SplitSeq(str, ",")

	if exists {
		for search := range searchList {
			search = strings.TrimSpace(search)

			s = append(s, search)
		}

		return s, value
	}

	return []string{}, part
}

func ParseTag(tag string) map[string]string {
	out := map[string]string{}
	parts := strings.SplitSeq(tag, "|")

	for part := range parts {
		keys, value := parseTagPart(part)

		if len(keys) == 0 {
			out["*"] = value
		}

		for _, key := range keys {
			out[key] = value
		}
	}

	return out
}