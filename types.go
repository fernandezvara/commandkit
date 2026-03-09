// commandkit/types.go
package commandkit

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValueType represents the expected type of a configuration value
type ValueType int

const (
	TypeString ValueType = iota
	TypeInt64
	TypeInt // Platform-specific int type
	TypeFloat64
	TypeBool
	TypeDuration
	TypeURL
	TypeStringSlice
	TypeInt64Slice
	TypeIntSlice // Platform-specific int slice type
	TypeUint
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeFloat32
	TypeTime
	TypeFloat64Slice
	TypeBoolSlice
	TypeFileMode
	TypeIP
	TypeUUID
	TypePath
)

func (t ValueType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInt64:
		return "int64"
	case TypeInt: // Platform-specific int type
		return "int"
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
	case TypeIntSlice: // Platform-specific int slice type
		return "[]int"
	case TypeUint:
		return "uint"
	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	case TypeUint64:
		return "uint64"
	case TypeFloat32:
		return "float32"
	case TypeTime:
		return "time.Time"
	case TypeFloat64Slice:
		return "[]float64"
	case TypeBoolSlice:
		return "[]bool"
	case TypeFileMode:
		return "filemode"
	case TypeIP:
		return "ip"
	case TypeUUID:
		return "uuid"
	case TypePath:
		return "path"
	default:
		return "unknown"
	}
}

// SourceType represents a configuration source type
type SourceType int

const (
	SourceDefault SourceType = iota
	SourceFlag
	SourceEnv
	SourceFile
)

func (s SourceType) String() string {
	switch s {
	case SourceDefault:
		return "default"
	case SourceFlag:
		return "flag"
	case SourceEnv:
		return "environment" // Changed from "env" to "environment" to match test expectations
	case SourceFile:
		return "file"
	default:
		return "unknown"
	}
}

// SourcePriority defines the priority order for configuration sources
type SourcePriority []SourceType

// Common priority presets
var (
	// PriorityFlagEnvDefault sets Flag > Env > Default priority
	PriorityFlagEnvDefault = SourcePriority{SourceFlag, SourceEnv, SourceDefault}

	// PriorityEnvFlagDefault sets Env > Flag > Default priority
	PriorityEnvFlagDefault = SourcePriority{SourceEnv, SourceFlag, SourceDefault}

	// PriorityFileEnvFlagDefault sets File > Env > Flag > Default priority (current default)
	PriorityFileEnvFlagDefault = SourcePriority{SourceFile, SourceEnv, SourceFlag, SourceDefault}

	// PriorityDefaultOnly uses only Default values
	PriorityDefaultOnly = SourcePriority{SourceDefault}
)

// expandPath expands path to full absolute path
func expandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	// Expand environment variables
	expanded := os.ExpandEnv(path)

	// Expand ~ to user home directory
	if strings.HasPrefix(expanded, "~") {
		if len(expanded) == 1 {
			// Just "~" - use current user home
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			expanded = home
		} else if expanded[1] == '/' || expanded[1] == filepath.Separator {
			// "~/path" - use current user home
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			expanded = filepath.Join(home, expanded[2:])
		} else {
			// "~user/path" - use specific user home
			username := expanded[1:]
			if slashIdx := strings.Index(username, "/"); slashIdx != -1 {
				username = username[:slashIdx]
				restPath := expanded[1+slashIdx:]

				usr, err := user.Lookup(username)
				if err != nil {
					return "", fmt.Errorf("user %s not found: %w", username, err)
				}
				expanded = filepath.Join(usr.HomeDir, restPath)
			} else {
				usr, err := user.Lookup(username)
				if err != nil {
					return "", fmt.Errorf("user %s not found: %w", username, err)
				}
				expanded = usr.HomeDir
			}
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Clean the path (resolve ., .., etc.)
	cleanedPath := filepath.Clean(absPath)

	return cleanedPath, nil
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

	case TypeInt:
		v, err := strconv.ParseInt(raw, 10, 64) // Parse as int64, store as int
		if err != nil {
			return nil, fmt.Errorf("invalid int: %s", raw)
		}
		return int(v), nil // Convert to int

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

	case TypeIntSlice:
		if raw == "" {
			return []int{}, nil
		}
		parts := strings.Split(raw, delimiter)
		result := make([]int, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			v, err := strconv.ParseInt(trimmed, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid int in array: %s", trimmed)
			}
			result = append(result, int(v)) // Convert to int
		}
		return result, nil

	// Unsigned integer types
	case TypeUint:
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint: %s", raw)
		}
		return uint(v), nil

	case TypeUint8:
		v, err := strconv.ParseUint(raw, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid uint8: %s", raw)
		}
		return uint8(v), nil

	case TypeUint16:
		v, err := strconv.ParseUint(raw, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid uint16: %s", raw)
		}
		return uint16(v), nil

	case TypeUint32:
		v, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid uint32: %s", raw)
		}
		return uint32(v), nil

	case TypeUint64:
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint64: %s", raw)
		}
		return v, nil

	case TypeFloat32:
		v, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid float32: %s", raw)
		}
		return float32(v), nil

	case TypeTime:
		// Try multiple time formats
		timeFormats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
			time.UnixDate,
			time.RubyDate,
		}

		// Try Unix timestamp first
		if timestamp, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return time.Unix(timestamp, 0), nil
		}

		// Try each format
		for _, format := range timeFormats {
			if t, err := time.Parse(format, raw); err == nil {
				return t, nil
			}
		}

		return nil, fmt.Errorf("invalid time format: %s (try RFC3339, Unix timestamp, or YYYY-MM-DD HH:MM:SS)", raw)

	case TypeFloat64Slice:
		if raw == "" {
			return []float64{}, nil
		}
		parts := strings.Split(raw, delimiter)
		result := make([]float64, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			v, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float64 in array: %s", trimmed)
			}
			result = append(result, v)
		}
		return result, nil

	case TypeBoolSlice:
		if raw == "" {
			return []bool{}, nil
		}
		parts := strings.Split(raw, delimiter)
		result := make([]bool, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			v, err := strconv.ParseBool(trimmed)
			if err != nil {
				return nil, fmt.Errorf("invalid bool in array: %s", trimmed)
			}
			result = append(result, v)
		}
		return result, nil

	case TypeFileMode:
		// Handle octal formats: "0755", "644", "0o755"
		var rawValue string
		if strings.HasPrefix(raw, "0o") || strings.HasPrefix(raw, "0O") {
			rawValue = raw[2:]
		} else if strings.HasPrefix(raw, "0") && len(raw) > 1 {
			rawValue = raw[1:]
		} else {
			rawValue = raw
		}

		value, err := strconv.ParseUint(rawValue, 8, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid file mode: %s", raw)
		}

		// Validate range for 32-bit file permissions
		if value > 0xFFFFFFFF {
			return nil, fmt.Errorf("file mode %s exceeds maximum 0xFFFFFFFF", raw)
		}

		return os.FileMode(uint32(value)), nil // Store as os.FileMode

	case TypeIP:
		if raw == "" {
			return nil, fmt.Errorf("IP address cannot be empty")
		}
		ip := net.ParseIP(raw)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address: %s", raw)
		}
		return raw, nil // Store as string for flexibility

	case TypeUUID:
		if raw == "" {
			return nil, fmt.Errorf("UUID cannot be empty")
		}
		// Use UUID library for proper parsing and validation
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %s", raw)
		}
		return parsed, nil // Store as uuid.UUID

	case TypePath:
		// Expand path to full absolute path
		expandedPath, err := expandPath(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}

		return expandedPath, nil // Store as expanded string

	default:
		return nil, fmt.Errorf("unknown type: %v", valueType)
	}
}
