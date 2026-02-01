package configutils

import (
	"fmt"
	"reflect"
)

type Opt[T any] struct {
	Set		bool
	Value	T
}

// Returns optional.Value (if set) or fallback
func (optional Opt[T]) ValueOrFallback(fallback T) T {
    if optional.Set {
        return optional.Value
    }

    return fallback
}

// Returns optional.Value (if set) or fallback.Value
func (optional Opt[T]) OptOrFallback(fallback Opt[T]) T {
    if optional.Set {
        return optional.Value
    }

    return fallback.Value
}

// Returns optional.Value (if set) or fallback.Value (if set), else T empty is returned
func (optional Opt[T]) OptOrEmpty(fallback Opt[T]) T {
    if optional.Set {
        return optional.Value
    }

    if fallback.Set {
        return fallback.Value
    }

    var zero T
    return zero
}

func (optional *Opt[T]) DecodeHook(raw any) error {
    value := reflect.ValueOf(raw)
    valueType := reflect.ValueOf(&optional.Value).Elem()

    if value.Type().ConvertibleTo(valueType.Type()) {
        valueType.Set(value.Convert(valueType.Type()))
        optional.Set = true
        return nil
    }
    return fmt.Errorf("cannot decode %v into Opt[%T]", raw, optional.Value)
}