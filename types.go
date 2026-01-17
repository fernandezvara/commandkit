// commandkit/types.go
package commandkit

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ValueType represents the expected type of a configuration value
type ValueType int

const (
	TypeString ValueType = iota
	TypeInt64
	TypeFloat64
	TypeBool
	TypeDuration
	TypeURL
	TypeStringSlice
	TypeInt64Slice
)

func (t ValueType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInt64:
		return "int64"
	case TypeFloat64:
		return "float64"
	case TypeBool:
		return "bool"
	case TypeDuration:
		return "duration"
	case TypeURL:
		return "url"
	case TypeStringSlice:
		return "[]string"
	case TypeInt64Slice:
		return "[]int64"
	default:
		return "unknown"
	}
}

// parseValue parses a string value into the expected type
func parseValue(raw string, valueType ValueType, delimiter string) (any, error) {
	if raw == "" {
		return nil, nil
	}

	switch valueType {
	case TypeString:
		return raw, nil

	case TypeInt64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int64: %s", raw)
		}
		return v, nil

	case TypeFloat64:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float64: %s", raw)
		}
		return v, nil

	case TypeBool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid bool: %s (use true/false, 1/0, yes/no)", raw)
		}
		return v, nil

	case TypeDuration:
		// Handle day format (e.g., "7d", "1d")
		if strings.HasSuffix(raw, "d") {
			daysStr := strings.TrimSuffix(raw, "d")
			if days, err := strconv.ParseFloat(daysStr, 64); err == nil {
				hours := days * 24
				v, err := time.ParseDuration(fmt.Sprintf("%.0fh", hours))
				if err != nil {
					return nil, fmt.Errorf("invalid duration: %s", raw)
				}
				return v, nil
			}
		}
		v, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: %s (use format like 15m, 1h, 7d)", raw)
		}
		return v, nil

	case TypeURL:
		v, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %s", raw)
		}
		if v.Scheme == "" || v.Host == "" {
			return nil, fmt.Errorf("invalid URL (missing scheme or host): %s", raw)
		}
		return raw, nil // Store as string, validated

	case TypeStringSlice:
		if raw == "" {
			return []string{}, nil
		}
		parts := strings.Split(raw, delimiter)
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil

	case TypeInt64Slice:
		if raw == "" {
			return []int64{}, nil
		}
		parts := strings.Split(raw, delimiter)
		result := make([]int64, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			v, err := strconv.ParseInt(trimmed, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid int64 in array: %s", trimmed)
			}
			result = append(result, v)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown type: %v", valueType)
	}
}
