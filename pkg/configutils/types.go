package configutils

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
)

type NullSentinel struct{}

func NullSentinelHook(_, _ reflect.Type, data any) (any, error) {
    if data == nil {
        return NullSentinel{}, nil
    }
    return data, nil
}

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

func (optional *Opt[T]) UnmarshalMapstructure(raw any) error {
    optional.Set = true

    _, ok := raw.(NullSentinel)
    if ok {
        // explici null
        return nil
    }

    return mapstructure.Decode(raw, &optional.Value)
}