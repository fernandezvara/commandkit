// commandkit/type_converter.go
package commandkit

import (
	"fmt"
	"strings"
)

// TypeConverter handles type-to-string conversions consistently across all value sources
type TypeConverter struct{}

// NewTypeConverter creates a new TypeConverter instance
func NewTypeConverter() *TypeConverter {
	return &TypeConverter{}
}

// ConvertToString converts any value to string representation for processing
// Returns error for unsupported types to ensure proper error handling
func (tc *TypeConverter) ConvertToString(value any, delimiter string) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case bool, int, int64, float64:
		return fmt.Sprintf("%v", v), nil
	case []string:
		// Handle string slices - join with delimiter
		return strings.Join(v, delimiter), nil
	case []int64:
		// Handle int64 slices - convert to strings and join
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%d", item)
		}
		return strings.Join(strs, delimiter), nil
	case []int:
		// Handle int slices - convert to strings and join
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%d", item)
		}
		return strings.Join(strs, delimiter), nil
	case []any:
		// Handle arrays from files - convert to strings and join
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(strs, delimiter), nil
	default:
		return "", fmt.Errorf("unsupported value type: %T", v)
	}
}

// ConvertToDisplayString converts value to string for display purposes
// Falls back to fmt.Sprintf for unsupported types to ensure display always works
func (tc *TypeConverter) ConvertToDisplayString(value any, delimiter string) string {
	str, err := tc.ConvertToString(value, delimiter)
	if err != nil {
		// For display purposes, fall back to basic string conversion
		return fmt.Sprintf("%v", value)
	}
	return str
}

// IsSupportedType checks if a type is supported for conversion
func (tc *TypeConverter) IsSupportedType(value any) bool {
	switch value.(type) {
	case string, bool, int, int64, float64:
		return true
	case []string, []int64, []int, []any:
		return true
	default:
		return false
	}
}
