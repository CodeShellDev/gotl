package configutils

import (
	"fmt"
	"maps"
	"reflect"
	"strconv"
	"strings"

	"github.com/codeshelldev/gotl/pkg/jsonutils"
)

type TransformTarget struct {
	Key string
	Transform string
	ChildTransform string
	Value any
}

// Apply Transform funcs based on `transform` and `childtransform` in struct schema
func (config Config) ApplyTransformFuncs(id string, structSchema any, path string, funcs map[string]func(string, any) (string, any)) {
	transformTargets := getKeyToTransformMap(id, structSchema)

	fmt.Println(id + ":\n" + jsonutils.Pretty(transformTargets))

	data := config.Layer.Get(path)

	_, res := applyTransform("", data, transformTargets, funcs)

	mapRes, ok := res.(map[string]any)

	if !ok {
		return
	}

	config.Layer.Delete("")
	config.Load(mapRes, path)
}

func getKeyToTransformMap(id string, value any) map[string]TransformTarget {
	data := map[string]TransformTarget{}

	if value == nil {
		return data
	}

	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return data
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		keys := []string{}

		outputKey := field.Tag.Get("koanf")

		keys = append(keys, outputKey)

		aliases := strings.Split(getFieldWithID(id, "aliases", field.Tag), ",")
		keys = append(keys, aliases...)

		for _, key := range keys {
			if key == "" {
				continue
			}

			lower := strings.ToLower(key)

			transformTag := getFieldWithID(id, "transform", field.Tag)
			childTransformTag := getFieldWithID(id, "childtransform", field.Tag)

			data[lower] = TransformTarget{
				Key:               strings.ToLower(outputKey), // Use `outputKey` here for aliasing
				Transform:         transformTag,
				ChildTransform:    childTransformTag,
				Value:             getValueSafe(fieldValue),
			}

			// Recursively walk nested structs
			if fieldValue.Kind() == reflect.Struct || (fieldValue.Kind() == reflect.Ptr && fieldValue.Elem().Kind() == reflect.Struct) {

				sub := getKeyToTransformMap(id, fieldValue.Interface())

				for subKey, subValue := range sub {
					if subKey == "" {
						continue
					}

					// `.` suffix means absolute (mainly useful for aliases)					
					if !strings.HasPrefix(subKey, ".") {
						subKey = lower + "." + strings.ToLower(subKey)
					} else {
						subKey = strings.ToLower(subKey[1:])
					}

					data[subKey] = subValue
				}
			}
		}
	}

	return data
}

func getValueSafe(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		return getValueSafe(value.Elem())
	}
	return value.Interface()
}

func getFieldWithID(id string, key string, tag reflect.StructTag) string {
	field, exists := tag.Lookup(id + ">" + key)

	if !exists {
		return tag.Get(key)
	}

	return field
}

func applyTransform(key string, value any, transformTargets map[string]TransformTarget, funcs map[string]func(string, any) (string, any)) (string, any) {
	lower := strings.ToLower(key)
	target := transformTargets[lower]

	targets := map[string]TransformTarget{}
		
	maps.Copy(targets, transformTargets)

	newKey, _ := applyTransformToAny(lower, value, transformTargets, funcs)

	newKeyWithDot := newKey

	if newKey != "" {
		newKeyWithDot = newKey + "."
	}

	switch asserted := value.(type) {
	case map[string]any:
		res := map[string]any{}

		for k, v := range asserted {
			fullKey := newKeyWithDot + k

			target, ok := targets[fullKey]

			fullKey = newKeyWithDot + target.Key

			if !ok {
				childTarget := TransformTarget{
					Key: fullKey,
					Transform: target.ChildTransform,
					ChildTransform: target.ChildTransform,
				}

				targets[fullKey] = childTarget
			}

			childKey, childValue := applyTransform(fullKey, v, targets, funcs)

			keyParts := getKeyParts(childKey)

			res[keyParts[len(keyParts)-1]] = childValue
		}

		keyParts := getKeyParts(newKey)

		return keyParts[len(keyParts)-1], res
	case []any:
		res := []any{}
		
		for i, child := range asserted {
			fullKey := newKeyWithDot + strconv.Itoa(i)

			_, ok := targets[fullKey]

			if !ok {
				childTarget := TransformTarget{
					Key: fullKey,
					Transform: target.ChildTransform,
					ChildTransform: target.ChildTransform,
				}

				targets[fullKey] = childTarget
			}
			
			_, childValue := applyTransform(fullKey, child, targets, funcs)

			res = append(res, childValue)
		}

		keyParts := getKeyParts(newKey)

		return keyParts[len(keyParts)-1], res
	default:
		return applyTransformToAny(key, asserted, transformTargets, funcs)
	}
}

func applyTransformToAny(key string, value any, transformTargets map[string]TransformTarget, funcs map[string]func(string, any) (string, any)) (string, any) {
	lower := strings.ToLower(key)

	transformTarget, ok := transformTargets[lower]
	if !ok {
		transformTarget.Transform = "default"
	}

	transformFuncs := strings.Split(transformTarget.Transform, ",")

 	resKey := key
	resValue := value

	for _, fnKey := range transformFuncs {
		fn, ok := funcs[fnKey]

		if !ok {
			fn = funcs["default"]
		}

		keyParts := getKeyParts(resKey)

		resKey, resValue = fn(keyParts[len(keyParts)-1], resValue)
	}

	return resKey, resValue
}

func getKeyParts(fullKey string) []string {
	keyParts := strings.Split(fullKey, ".")

	return keyParts
}