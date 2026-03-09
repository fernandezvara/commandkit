// commandkit/type_converter.go
package commandkit

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
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
	case uint, uint8, uint16, uint32:
		return fmt.Sprintf("%v", v), nil
	case uint64:
		return fmt.Sprintf("%d", v), nil
	case float32:
		return fmt.Sprintf("%g", v), nil
	case time.Time:
		return v.Format(time.RFC3339), nil
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
	case []float64:
		// Handle float64 slices - convert to strings and join
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%g", item)
		}
		return strings.Join(strs, delimiter), nil
	case []bool:
		// Handle bool slices - convert to strings and join
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%t", item)
		}
		return strings.Join(strs, delimiter), nil
	case []any:
		// Handle arrays from files - convert to strings and join
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(strs, delimiter), nil
	case os.FileMode:
		// Display FileMode in octal format
		return fmt.Sprintf("0%o", v), nil
	case net.IP:
		// Display IP address in standard format
		return v.String(), nil
	case uuid.UUID:
		// Display UUID in standard format
		return v.String(), nil
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
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return true
	case time.Time:
		return true
	case []string, []int64, []int, []any:
		return true
	case []float64, []bool:
		return true
	case os.FileMode:
		return true
	case net.IP:
		return true
	case uuid.UUID:
		return true
	default:
		return false
	}
}
