package configutils

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
)

type NilSentinel struct{}

func NilSentinelHook(_, _ reflect.Type, data any) (any, error) {
    if data == nil {
        return NilSentinel{}, nil
    }
    return data, nil
}

type Opt[T any] struct {
	set		bool
	value	*T
}

func (optional Opt[T]) Value() T {
    return *optional.value
}

func (optional Opt[T]) Set() bool {
    return optional.set
}

// Returns optional.Value (if set) or fallback
func (optional Opt[T]) ValueOrFallback(fallback T) T {
    if optional.set {
        return *optional.value
    }

    return fallback
}

// Returns optional.Value (if set) or fallback.Value
func (optional Opt[T]) OptOrFallback(fallback Opt[T]) T {
    if optional.set {
        return *optional.value
    }

    return *fallback.value
}

// Returns optional.Value (if set) or fallback.Value (if set), else T empty is returned
func (optional Opt[T]) OptOrEmpty(fallback Opt[T]) T {
    if optional.set {
        return *optional.value
    }

    if fallback.set {
        return *fallback.value
    }

    var zero T
    return zero
}

func (optional *Opt[T]) UnmarshalMapstructure(raw any) error {
    optional.set = true

    _, ok := raw.(NilSentinel)
    if ok {
        // explicit null
        return nil
    }

    var value T

    err := mapstructure.Decode(raw, &value)
    if err != nil {
        return err
    }

    optional.value = &value

    return nil
}