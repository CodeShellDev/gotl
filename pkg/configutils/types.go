package configutils

import (
	"github.com/go-viper/mapstructure/v2"
)

type Opt[T any] struct {
	Set		bool
	Value	T
}

func (optional Opt[T]) ValueOr(fallback T) T {
    if optional.Set {
        return optional.Value
    }

    return fallback
}

func (optional *Opt[T]) UnmarshalMapstructure(raw any) error {
    optional.Set = true

    return mapstructure.Decode(raw, &optional.Value)
}