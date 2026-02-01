package configutils

import (
	"github.com/go-viper/mapstructure/v2"
)

type Optional[T any] struct {
	Set		bool
	Value	T
}

func (optional Optional[T]) ValueOr(fallback T) T {
    if optional.Set {
        return optional.Value
    }

    return fallback
}

func (optional *Optional[T]) UnmarshalMapstructure(raw any) error {
    optional.Set = true

    return mapstructure.Decode(raw, &optional.Value)
}