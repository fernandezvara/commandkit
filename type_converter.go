// commandkit/type_converter.go
package commandkit

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
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

// convertDefaultValue handles type conversion for default values based on target ValueType
func convertDefaultValue(value any, targetType ValueType) (any, error) {
	if value == nil {
		return nil, nil
	}

	// If value is already the correct type, return it directly
	switch targetType {
	case TypeString:
		if _, ok := value.(string); ok {
			return value, nil
		}
	case TypeInt64:
		if _, ok := value.(int64); ok {
			return value, nil
		}
	case TypeInt:
		if _, ok := value.(int); ok {
			return value, nil
		}
	case TypeFloat64:
		if _, ok := value.(float64); ok {
			return value, nil
		}
	case TypeBool:
		if _, ok := value.(bool); ok {
			return value, nil
		}
	case TypeDuration:
		if _, ok := value.(time.Duration); ok {
			return value, nil
		}
		// Add more direct type checks as needed
	}

	// Perform type conversions based on target type
	switch targetType {
	case TypeString:
		return fmt.Sprintf("%v", value), nil

	case TypeInt64:
		switch v := value.(type) {
		case int:
			return int64(v), nil
		case int32:
			return int64(v), nil
		case int16:
			return int64(v), nil
		case int8:
			return int64(v), nil
		case uint:
			if uint64(v) > uint64(1<<63-1) {
				return nil, fmt.Errorf("value %d overflows int64", v)
			}
			return int64(v), nil
		case uint32:
			return int64(v), nil
		case uint16:
			return int64(v), nil
		case uint8:
			return int64(v), nil
		case float64:
			if v > float64(1<<63-1) || v < float64(-1<<63) {
				return nil, fmt.Errorf("value %f overflows int64", v)
			}
			return int64(v), nil
		case float32:
			if v > float32(1<<63-1) || v < float32(-1<<63) {
				return nil, fmt.Errorf("value %f overflows int64", v)
			}
			return int64(v), nil
		case string:
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to int64: %w", v, err)
			}
			return parsed, nil
		case bool:
			if v {
				return int64(1), nil
			}
			return int64(0), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to int64", value)
		}

	case TypeInt:
		switch v := value.(type) {
		case int64:
			if v > int64(1<<31-1) || v < int64(-1<<31) {
				return nil, fmt.Errorf("value %d overflows int on this platform", v)
			}
			return int(v), nil
		case int32:
			return int(v), nil
		case int16:
			return int(v), nil
		case int8:
			return int(v), nil
		case uint:
			if int64(v) > int64(1<<31-1) {
				return nil, fmt.Errorf("value %d overflows int on this platform", v)
			}
			return int(v), nil
		case uint32:
			return int(v), nil
		case uint16:
			return int(v), nil
		case uint8:
			return int(v), nil
		case float64:
			if v > float64(1<<31-1) || v < float64(-1<<31) {
				return nil, fmt.Errorf("value %f overflows int on this platform", v)
			}
			return int(v), nil
		case float32:
			if v > float32(1<<31-1) || v < float32(-1<<31) {
				return nil, fmt.Errorf("value %f overflows int on this platform", v)
			}
			return int(v), nil
		case string:
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to int: %w", v, err)
			}
			if parsed > int64(1<<31-1) || parsed < int64(-1<<31) {
				return nil, fmt.Errorf("value %s overflows int on this platform", v)
			}
			return int(parsed), nil
		case bool:
			if v {
				return int(1), nil
			}
			return int(0), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to int", value)
		}

	case TypeFloat64:
		switch v := value.(type) {
		case int, int8, int16, int32, int64:
			return float64(reflect.ValueOf(v).Int()), nil
		case uint, uint8, uint16, uint32, uint64:
			return float64(reflect.ValueOf(v).Uint()), nil
		case float32:
			return float64(v), nil
		case string:
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to float64: %w", v, err)
			}
			return parsed, nil
		case bool:
			if v {
				return float64(1), nil
			}
			return float64(0), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to float64", value)
		}

	case TypeBool:
		switch v := value.(type) {
		case string:
			lower := strings.ToLower(v)
			switch lower {
			case "true", "1", "yes", "on", "enabled":
				return true, nil
			case "false", "0", "no", "off", "disabled":
				return false, nil
			default:
				return nil, fmt.Errorf("cannot convert string '%s' to bool", v)
			}
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(v).Int() != 0, nil
		case uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(v).Uint() != 0, nil
		case float32, float64:
			return reflect.ValueOf(v).Float() != 0, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to bool", value)
		}

	case TypeDuration:
		switch v := value.(type) {
		case string:
			parsed, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to duration: %w", v, err)
			}
			return parsed, nil
		case int, int8, int16, int32, int64:
			// Assume seconds for integer values
			seconds := reflect.ValueOf(v).Int()
			return time.Duration(seconds) * time.Second, nil
		case float32, float64:
			// Assume seconds for float values
			seconds := reflect.ValueOf(v).Float()
			return time.Duration(seconds) * time.Second, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to duration", value)
		}

	// Handle slice types
	case TypeStringSlice:
		switch v := value.(type) {
		case []string:
			return v, nil
		case []int:
			result := make([]string, len(v))
			for i, item := range v {
				result[i] = fmt.Sprintf("%d", item)
			}
			return result, nil
		case []int64:
			result := make([]string, len(v))
			for i, item := range v {
				result[i] = fmt.Sprintf("%d", item)
			}
			return result, nil
		case []bool:
			result := make([]string, len(v))
			for i, item := range v {
				result[i] = fmt.Sprintf("%t", item)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to []string", value)
		}

	case TypeInt64Slice:
		switch v := value.(type) {
		case []int64:
			return v, nil
		case []int:
			result := make([]int64, len(v))
			for i, item := range v {
				result[i] = int64(item)
			}
			return result, nil
		case []string:
			result := make([]int64, len(v))
			for i, item := range v {
				parsed, err := strconv.ParseInt(item, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot convert string '%s' to int64 in slice: %w", item, err)
				}
				result[i] = parsed
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to []int64", value)
		}

	case TypeIntSlice:
		switch v := value.(type) {
		case []int:
			return v, nil
		case []int64:
			result := make([]int, len(v))
			for i, item := range v {
				if item > int64(1<<31-1) || item < int64(-1<<31) {
					return nil, fmt.Errorf("value %d overflows int on this platform in slice", item)
				}
				result[i] = int(item)
			}
			return result, nil
		case []string:
			result := make([]int, len(v))
			for i, item := range v {
				parsed, err := strconv.ParseInt(item, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot convert string '%s' to int in slice: %w", item, err)
				}
				if parsed > int64(1<<31-1) || parsed < int64(-1<<31) {
					return nil, fmt.Errorf("value %s overflows int on this platform in slice", item)
				}
				result[i] = int(parsed)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to []int", value)
		}

	// Handle unsigned integer types
	case TypeUint:
		switch v := value.(type) {
		case uint:
			return v, nil
		case int:
			if v < 0 {
				return nil, fmt.Errorf("cannot convert negative int %d to uint", v)
			}
			return uint(v), nil
		case int64:
			if v < 0 {
				return nil, fmt.Errorf("cannot convert negative int64 %d to uint", v)
			}
			return uint(v), nil
		case string:
			parsed, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to uint: %w", v, err)
			}
			return uint(parsed), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to uint", value)
		}

	case TypeUint8:
		switch v := value.(type) {
		case uint8:
			return v, nil
		case int:
			if v < 0 || v > 255 {
				return nil, fmt.Errorf("value %d out of range for uint8", v)
			}
			return uint8(v), nil
		case int64:
			if v < 0 || v > 255 {
				return nil, fmt.Errorf("value %d out of range for uint8", v)
			}
			return uint8(v), nil
		case string:
			parsed, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to uint8: %w", v, err)
			}
			return uint8(parsed), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to uint8", value)
		}

	case TypeUint16:
		switch v := value.(type) {
		case uint16:
			return v, nil
		case int:
			if v < 0 || v > 65535 {
				return nil, fmt.Errorf("value %d out of range for uint16", v)
			}
			return uint16(v), nil
		case int64:
			if v < 0 || v > 65535 {
				return nil, fmt.Errorf("value %d out of range for uint16", v)
			}
			return uint16(v), nil
		case string:
			parsed, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to uint16: %w", v, err)
			}
			return uint16(parsed), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to uint16", value)
		}

	case TypeUint32:
		switch v := value.(type) {
		case uint32:
			return v, nil
		case int:
			if v < 0 || v > 4294967295 {
				return nil, fmt.Errorf("value %d out of range for uint32", v)
			}
			return uint32(v), nil
		case int64:
			if v < 0 || v > 4294967295 {
				return nil, fmt.Errorf("value %d out of range for uint32", v)
			}
			return uint32(v), nil
		case string:
			parsed, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to uint32: %w", v, err)
			}
			return uint32(parsed), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to uint32", value)
		}

	case TypeUint64:
		switch v := value.(type) {
		case uint64:
			return v, nil
		case int:
			if v < 0 {
				return nil, fmt.Errorf("cannot convert negative int %d to uint64", v)
			}
			return uint64(v), nil
		case int64:
			if v < 0 {
				return nil, fmt.Errorf("cannot convert negative int64 %d to uint64", v)
			}
			return uint64(v), nil
		case string:
			parsed, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to uint64: %w", v, err)
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to uint64", value)
		}

	case TypeFloat32:
		switch v := value.(type) {
		case float32:
			return v, nil
		case int, int8, int16, int32, int64:
			return float32(reflect.ValueOf(v).Int()), nil
		case uint, uint8, uint16, uint32, uint64:
			return float32(reflect.ValueOf(v).Uint()), nil
		case float64:
			return float32(v), nil
		case bool:
			if v {
				return float32(1), nil
			}
			return float32(0), nil
		case string:
			parsed, err := strconv.ParseFloat(v, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to float32: %w", v, err)
			}
			return float32(parsed), nil
		default:
			return nil, fmt.Errorf("cannot convert %T to float32", value)
		}

	// Add more type conversions as needed for other ValueType constants
	default:
		return value, nil // No conversion needed for unsupported types
	}
}

// ConvertDefaultValue converts a default value to the target ValueType
// This is the main method that should be used for default value type conversion
func (tc *TypeConverter) ConvertDefaultValue(value any, targetType ValueType) (any, error) {
	return convertDefaultValue(value, targetType)
}
