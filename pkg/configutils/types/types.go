package configutils

import (
	"errors"
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
	Set		bool
	Value	*T
}

// Returns optional.Value (if set) or fallback
func (optional Opt[T]) ValueOrFallback(fallback T) T {
    if optional.Set {
        if optional.Value != nil {
            return *optional.Value
        }

        var zero T
        return zero
    }

    return fallback
}

// Returns optional.Value (if set) or fallback.Value
func (optional Opt[T]) OptOrFallback(fallback Opt[T]) T {
    if optional.Set {
        if optional.Value != nil {
            return *optional.Value
        }

        var zero T
        return zero
    }

    if fallback.Value != nil {
        if fallback.Value != nil {
            return *fallback.Value
        }

        var zero T
        return zero
    }

    var zero T
    return zero
}

// Returns optional.Value (if set) or fallback.Value (if set), else T empty is returned
func (optional Opt[T]) OptOrEmpty(fallback Opt[T]) T {
    if optional.Set {
        if optional.Value != nil {
            return *optional.Value
        }

        var zero T
        return zero
    }

    if fallback.Set {
        if fallback.Value != nil {
            return *fallback.Value
        }

        var zero T
        return zero
    }

    var zero T
    return zero
}

func (optional *Opt[T]) UnmarshalMapstructure(raw any) error {
    optional.Set = true

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

    optional.Value = &value

    return nil
}

type Compilable[CompiledT any] interface {
	Compile() CompiledT
}

type Comp[RawT Compilable[CompiledT], CompiledT any] struct {
	Raw        *RawT
	compiled   *CompiledT
	done bool
}

func (c *Comp[RawT, CompiledT]) Compile() CompiledT {
    if c == nil {
        var zero CompiledT
        return zero
    }

	if c.done {
		return *c.compiled
	}

	compiled := (*c.Raw).Compile()

	c.compiled = &compiled
	c.done = true

	return compiled
}

func (c *Comp[RawT, CompiledT]) UnmarshalMapstructure(raw any) error {
    if c == nil {
        return errors.New("compiled struct cannot be nil")
    }

	var rawT RawT

    err := mapstructure.Decode(raw, &rawT)
    
	if err != nil {
		return err
	}

	c.Raw = &rawT

	return nil
}
