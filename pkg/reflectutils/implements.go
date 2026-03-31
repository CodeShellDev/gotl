package reflectutils

import (
	"reflect"
)

func OnImplements[T any](current reflect.Type, targetInterface reflect.Type, fn func(iface T) bool) bool {
	if current == nil {
		return true
	}

	// unwrap pointers
	for current.Kind() == reflect.Pointer {
		current = current.Elem()
	}

	// check if implements interface
	if current.Implements(targetInterface) {
		value := reflect.New(current).Elem()

		iface := value.Interface().(T)

		if !fn(iface) {
			return false
		}
	}

	// also check pointer receiver case
	if reflect.PointerTo(current).Implements(targetInterface) {
		value := reflect.New(current)

		iface := value.Interface().(T)

		if !fn(iface) {
			return false
		}
	}

	switch current.Kind() {

	case reflect.Struct:
		for field := range current.Fields() {
			proceed := OnImplements(field.Type, targetInterface, fn)

			if !proceed {
				return false
			}
		}

	case reflect.Slice, reflect.Array:
		return OnImplements(current.Elem(), targetInterface, fn)

	case reflect.Map:
		proceed := OnImplements(current.Key(), targetInterface, fn)

		if !proceed {
			return false
		}

		proceed = OnImplements(current.Elem(), targetInterface, fn)

		if !proceed {
			return false
		}
	}

	return true
}