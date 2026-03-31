package reflectutils

import (
	"reflect"
)

func SafeInterface(value reflect.Value) (any, bool) {
    if !value.IsValid() {
        return nil, false
    }

    if value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
        if value.IsNil() {
            return nil, false
        }
        value = value.Elem()
    }

    if !value.CanInterface() {
        return nil, false
    }

    return value.Interface(), true
}

func OnTypeImplements[T any](current reflect.Type, targetInterface reflect.Type, fn func(iface T, t reflect.Type) bool) bool {
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

		if !fn(iface, current) {
			return false
		}
	}

	// also check pointer receiver case
	if reflect.PointerTo(current).Implements(targetInterface) {
		value := reflect.New(current)

		iface := value.Interface().(T)

		if !fn(iface, current) {
			return false
		}
	}

	switch current.Kind() {

	case reflect.Struct:
		for field := range current.Fields() {
			if !OnTypeImplements(field.Type, targetInterface, fn) {
				return false
			}
		}

	case reflect.Slice, reflect.Array:
		if !OnTypeImplements(current.Elem(), targetInterface, fn) {
			return false
		}

	case reflect.Map:
		if !OnTypeImplements(current.Key(), targetInterface, fn) {
			return false
		}

		if !OnTypeImplements(current.Elem(), targetInterface, fn) {
			return false
		}
	}

	return true
}

func OnValueImplements[T any](current reflect.Value, targetInterface reflect.Type, fn func(iface T, value reflect.Value) bool) bool {
	if !current.IsValid() {
		return true
	}

	// unwrap interface | pointer
	for current.Kind() == reflect.Interface || current.Kind() == reflect.Pointer {
		if current.IsNil() {
			return true
		}
		current = current.Elem()
	}

	t := current.Type()

	// check value implements interface
	if t.Implements(targetInterface) {
		iface := current.Interface().(T)

		if !fn(iface, current) {
			return false
		}
	}

	// check pointer receiver case
	if reflect.PointerTo(t).Implements(targetInterface) {
		if current.CanAddr() {
			iface := current.Addr().Interface().(T)
			
			if !fn(iface, current) {
				return false
			}
		}
	}

	switch current.Kind() {

	case reflect.Struct:
		for _, field := range current.Fields() {
			if !OnValueImplements(field, targetInterface, fn) {
				return false
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < current.Len(); i++ {
			if !OnValueImplements(current.Index(i), targetInterface, fn) {
				return false
			}
		}

	case reflect.Map:
		for _, k := range current.MapKeys() {
			if !OnValueImplements(current.MapIndex(k), targetInterface, fn) {
				return false
			}
		}
	}

	return true
}