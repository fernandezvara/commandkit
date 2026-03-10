// commandkit/get.go
package commandkit

import (
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// typeCache provides performance optimization for type descriptions
var typeCache = sync.Map{}

// GetError represents an error collected from Get functions
type GetError struct {
	Key              string
	ExpectedType     string
	ActualType       string
	Message          string
	IsSecret         bool
	Flag             string // Flag name (e.g., "port")
	EnvVar           string // Environment variable name (e.g., "PORT")
	Display          string
	ErrorDescription string
	config           *Config // Reference to config for definition lookup
}

// typeDescription returns cached type description for performance
func typeDescription(v any) string {
	typ := reflect.TypeOf(v)
	if cached, ok := typeCache.Load(typ); ok {
		return cached.(string)
	}

	desc := getTypeDescription(v)
	typeCache.Store(typ, desc)
	return desc
}

// getTypeDescription converts type to human-readable string
func getTypeDescription(v any) string {
	switch v.(type) {
	case string:
		return "string"
	case int64:
		return "int64"
	case int: // Handle platform-specific int type
		return "int"
	case bool:
		return "bool"
	case float64:
		return "float64"
	case []string:
		return "[]string"
	case []int64:
		return "[]int64"
	case []int: // Handle platform-specific int slice type
		return "[]int"
	case uint:
		return "uint"
	case uint8:
		return "uint8"
	case uint16:
		return "uint16"
	case uint32:
		return "uint32"
	case uint64:
		return "uint64"
	case float32:
		return "float32"
	case time.Time:
		return "time.Time"
	case []float64:
		return "[]float64"
	case []bool:
		return "[]bool"
	case os.FileMode:
		return "os.FileMode"
	case net.IP:
		return "net.IP"
	case uuid.UUID:
		return "uuid.UUID"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// newTypeError creates optimized type mismatch error
func newTypeError[T any](key string, value any) error {
	expectedType := typeDescription(*new(T))
	actualType := typeDescription(value)
	return fmt.Errorf("configuration '%s' type mismatch: expected %s, got %s",
		key, expectedType, actualType)
}

// convertValue handles type conversion for Get[T] compatibility
func convertValue(value any, targetType reflect.Type) (any, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil value")
	}

	sourceValue := reflect.ValueOf(value)
	sourceType := sourceValue.Type()

	// Handle specific type conversions first (before generic kind conversions)
	// Handle int to os.FileMode conversion (for default values)
	if targetType == reflect.TypeOf(os.FileMode(0)) && sourceType.Kind() == reflect.Int {
		return os.FileMode(sourceValue.Int()), nil
	}

	// Handle int to unsigned integer conversions (but not for specific types like FileMode)
	switch targetType.Kind() {
	case reflect.Uint:
		if sourceType.Kind() == reflect.Int {
			return uint(sourceValue.Int()), nil
		}
	case reflect.Uint8:
		if sourceType.Kind() == reflect.Int {
			return uint8(sourceValue.Int()), nil
		}
	case reflect.Uint16:
		if sourceType.Kind() == reflect.Int {
			return uint16(sourceValue.Int()), nil
		}
	case reflect.Uint32:
		if sourceType.Kind() == reflect.Int {
			return uint32(sourceValue.Int()), nil
		}
	case reflect.Uint64:
		if sourceType.Kind() == reflect.Int {
			return uint64(sourceValue.Int()), nil
		}
	}

	// FileMode is now stored directly as os.FileMode, no conversion needed

	// Handle string to net.IP conversion
	if targetType == reflect.TypeOf(net.IP{}) && sourceType == reflect.TypeOf("") {
		ip := net.ParseIP(sourceValue.String())
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address: %s", sourceValue.String())
		}
		return ip, nil
	}

	// Handle string to int64 conversion
	if targetType == reflect.TypeOf(int64(0)) && sourceType == reflect.TypeOf("") {
		parsed, err := strconv.ParseInt(sourceValue.String(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int64: %s", sourceValue.String())
		}
		return parsed, nil
	}

	// Handle string to uuid.UUID conversion
	if targetType == reflect.TypeOf(uuid.UUID{}) && sourceType == reflect.TypeOf("") {
		parsed, err := uuid.Parse(sourceValue.String())
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %s", sourceValue.String())
		}
		return parsed, nil
	}

	// Handle path expansion for string defaults
	if sourceType == reflect.TypeOf("") && sourceValue.String() != "" {
		pathStr := sourceValue.String()
		// Check if this looks like a path that needs expansion
		if strings.Contains(pathStr, "~") || strings.Contains(pathStr, "$") ||
			strings.HasPrefix(pathStr, "./") || strings.HasPrefix(pathStr, "../") {
			expanded, err := expandPath(pathStr)
			if err == nil {
				return expanded, nil
			}
		}
	}

	return value, nil
}

// logWarningForDesigner logs warnings for CLI designers about configuration issues
func logWarningForDesigner(message string) {
	// Use log package with a distinct prefix for designer warnings
	log.Printf("[CONFIG WARNING] %s", message)
}

// getErrorDisplayName returns the display name matching help message format
func getErrorDisplayName(err GetError, c *Config) string {
	if c != nil {
		if def, hasDef := c.definitions[err.Key]; hasDef {
			return buildErrorDisplay(def)
		}
	}

	if err.Flag != "" {
		return fmt.Sprintf("--%s string", err.Flag)
	}
	if err.EnvVar != "" {
		return fmt.Sprintf("(no flag) string (env: %s)", err.EnvVar)
	}
	return fmt.Sprintf("%s string", err.Key)
}

// Get retrieves a configuration value with proper error handling
// This function performs early secret detection to prevent type assertion exposure
func Get[T any](ctx *CommandContext, key string) (T, error) {
	var zero T

	// Determine which config to use (command config takes precedence)
	var c *Config
	if ctx.CommandConfig != nil {
		// Check if the key is defined in command config
		if _, hasDef := ctx.CommandConfig.definitions[key]; hasDef {
			c = ctx.CommandConfig
		} else {
			// Key not in command config, check global config
			c = ctx.GlobalConfig
		}
	} else {
		c = ctx.GlobalConfig
	}

	// Early secret detection - check definition before accessing values
	def, hasDef := c.definitions[key]
	if hasDef && def.secret {
		// Secure error handling - no secret exposure
		ctx.execution.CollectError(c, key, "secret", "", "use GetSecret() instead", true)
		result := validationError(fmt.Sprintf("configuration '%s' is secret, use GetSecret() instead", key))
		return zero, result.Error
	}

	value, exists := c.values[key]
	if !exists {
		// Check if this is required data - if so, return validation error
		if hasDef && def.required {
			// Log warning for designer and return validation error
			logWarningForDesigner(fmt.Sprintf("Required key '%s' is not provided", key))
			return zero, fmt.Errorf("required configuration '%s' not provided", key)
		}
		// For non-required keys, collect error and return result
		ctx.execution.CollectError(c, key, "not found", "", "key not defined", false)
		return zero, fmt.Errorf("configuration '%s' not found", key)
	}

	// Additional safety check - ensure value is not a secret placeholder
	if strVal, ok := value.(string); ok && strVal == "[SECRET]" {
		// This should not happen with new implementation, but add safety check
		ctx.execution.CollectError(c, key, "secret", "", "use GetSecret() instead", true)
		result := validationError(fmt.Sprintf("configuration '%s' is secret, use GetSecret() instead", key))
		return zero, result.Error
	}

	// Special handling for Path types - always expand paths
	if hasDef && def.valueType == TypePath {
		if strVal, ok := value.(string); ok {
			expanded, err := expandPath(strVal)
			if err != nil {
				return zero, fmt.Errorf("failed to expand path '%s': %w", strVal, err)
			}
			value = expanded
		}
	}

	// Special handling for FileMode types - parse string defaults
	if hasDef && def.valueType == TypeFileMode {
		if strVal, ok := value.(string); ok {
			parsed, err := parseValue(strVal, TypeFileMode, ",")
			if err != nil {
				return zero, fmt.Errorf("failed to parse file mode '%s': %w", strVal, err)
			}
			value = parsed
		}
	}

	// Special handling for IP types - validate string defaults
	if hasDef && def.valueType == TypeIP {
		if strVal, ok := value.(string); ok {
			if strVal == "" {
				return zero, fmt.Errorf("IP address cannot be empty")
			}
			parsed, err := parseValue(strVal, TypeIP, ",")
			if err != nil {
				return zero, fmt.Errorf("failed to parse IP address '%s': %w", strVal, err)
			}
			value = parsed
		}
	}

	// Special handling for UUID types - validate string defaults
	if hasDef && def.valueType == TypeUUID {
		if strVal, ok := value.(string); ok {
			if strVal == "" {
				return zero, fmt.Errorf("UUID cannot be empty")
			}
			parsed, err := parseValue(strVal, TypeUUID, ",")
			if err != nil {
				return zero, fmt.Errorf("failed to parse UUID '%s': %w", strVal, err)
			}
			value = parsed
		}
	}

	// Try direct type assertion first
	result, ok := value.(T)
	if ok {
		return result, nil
	}

	// Try type conversion for compatible types
	targetType := reflect.TypeOf(zero)
	converted, err := convertValue(value, targetType)
	if err == nil {
		result, ok = converted.(T)
		if ok {
			return result, nil
		}
	}

	// Type conversion failed
	ctx.execution.CollectError(c, key, typeDescription(*new(T)), typeDescription(value), "type mismatch", false)
	return zero, newTypeError[T](key, value)
}

// MustGet retrieves a configuration value and panics on error
// Use when you expect the configuration to be valid and want to fail fast
func MustGet[T any](ctx *CommandContext, key string) T {
	result, err := Get[T](ctx, key)
	if err != nil {
		panic(err)
	}
	return result
}

// Has checks if a key exists and has a non-nil value
// Note: This function will return false for secret keys to prevent exposure
func (c *Config) Has(key string) bool {
	// Check if this is a secret key - if so, always return false to prevent exposure
	if def, hasDef := c.definitions[key]; hasDef && def.secret {
		return false
	}

	value, exists := c.values[key]
	return exists && value != nil
}

// GetSecret retrieves a secret value securely
func (c *Config) GetSecret(key string) *Secret {
	return c.secrets.Get(key)
}

// HasSecret checks if a secret exists and is set
func (c *Config) HasSecret(key string) bool {
	return c.secrets.Has(key)
}

// Keys returns all defined configuration keys
func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.definitions))
	for k := range c.definitions {
		keys = append(keys, k)
	}
	return keys
}
