package config

import (
	"os"

	"github.com/spf13/viper"
)

func WithDefault[T any](key string, defaultValue T) T {
	viperReader := viper.GetViper()

	expandIfString := func(value T) T {
		asString, ok := any(value).(string)

		if !ok {
			return value
		}

		return any(os.ExpandEnv(asString)).(T)
	}

	if !viperReader.IsSet(key) {
		return expandIfString(defaultValue)
	}

	raw := viperReader.Get(key)

	if raw == nil {
		return expandIfString(defaultValue)
	}

	value, ok := raw.(T)

	if !ok {
		return expandIfString(defaultValue)
	}

	return expandIfString(value)
}
