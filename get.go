// commandkit/get.go
package commandkit

import (
	"fmt"
	"time"
)

// Get retrieves a configuration value with type safety using generics
func Get[T any](c *Config, key string) T {
	value, exists := c.values[key]
	if !exists {
		panic(fmt.Sprintf("commandkit: key '%s' not found (did you define it?)", key))
	}

	// Check if it's a secret (stored as string, needs special handling)
	def, hasDef := c.definitions[key]
	if hasDef && def.secret {
		panic(fmt.Sprintf("commandkit: key '%s' is a secret, use GetSecret() instead", key))
	}

	result, ok := value.(T)
	if !ok {
		panic(fmt.Sprintf("commandkit: key '%s' has type %T, not %T", key, value, result))
	}

	return result
}

// MustGet is an alias for Get (both panic on error)
func MustGet[T any](c *Config, key string) T {
	return Get[T](c, key)
}

// GetOr retrieves a configuration value or returns a default if not set
func GetOr[T any](c *Config, key string, defaultValue T) T {
	value, exists := c.values[key]
	if !exists || value == nil {
		return defaultValue
	}

	result, ok := value.(T)
	if !ok {
		return defaultValue
	}

	return result
}

// Has checks if a key exists and has a non-nil value
func (c *Config) Has(key string) bool {
	value, exists := c.values[key]
	return exists && value != nil
}

// GetSecret retrieves a secret value
func (c *Config) GetSecret(key string) *Secret {
	return c.secrets.Get(key)
}

// Keys returns all defined configuration keys
func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.definitions))
	for k := range c.definitions {
		keys = append(keys, k)
	}
	return keys
}

// Convenience typed getters (non-generic alternative)

func (c *Config) GetString(key string) string {
	return Get[string](c, key)
}

func (c *Config) GetInt64(key string) int64 {
	return Get[int64](c, key)
}

func (c *Config) GetFloat64(key string) float64 {
	return Get[float64](c, key)
}

func (c *Config) GetBool(key string) bool {
	return Get[bool](c, key)
}

func (c *Config) GetDuration(key string) time.Duration {
	return Get[time.Duration](c, key)
}

func (c *Config) GetStringSlice(key string) []string {
	return Get[[]string](c, key)
}

func (c *Config) GetInt64Slice(key string) []int64 {
	return Get[[]int64](c, key)
}
