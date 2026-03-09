// commandkit/validation.go
package commandkit

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"
	"time"
)

// ValidationCache stores pre-compiled regex patterns for performance
type ValidationCache struct {
	regexCache map[string]*regexp.Regexp
	mu         sync.RWMutex
}

var (
	// Global validation cache instance
	validationCache = &ValidationCache{
		regexCache: make(map[string]*regexp.Regexp),
	}
)

// getRegexp returns a compiled regex pattern, using cache for performance
func (vc *ValidationCache) getRegexp(pattern string) *regexp.Regexp {
	vc.mu.RLock()
	if re, ok := vc.regexCache[pattern]; ok {
		vc.mu.RUnlock()
		return re
	}
	vc.mu.RUnlock()

	// Compile and cache
	vc.mu.Lock()
	defer vc.mu.Unlock()
	// Double-check pattern
	re := regexp.MustCompile(pattern)
	vc.regexCache[pattern] = re
	return re
}

// Validation represents a validation rule
type Validation struct {
	Name  string
	Check func(value any) error
}

// Built-in validation constructors

func validateRequired() Validation {
	return Validation{
		Name: "required",
		Check: func(value any) error {
			if value == nil {
				return fmt.Errorf("value is required")
			}
			if s, ok := value.(string); ok && s == "" {
				return fmt.Errorf("value is required (empty string)")
			}
			return nil
		},
	}
}

func validateMin(min float64) Validation {
	return Validation{
		Name: fmt.Sprintf("min(%v)", min),
		Check: func(value any) error {
			switch v := value.(type) {
			case int64:
				if float64(v) < min {
					return fmt.Errorf("value %d is less than minimum %v", v, min)
				}
			case float64:
				if v < min {
					return fmt.Errorf("value %f is less than minimum %v", v, min)
				}
			}
			return nil
		},
	}
}

func validateMax(max float64) Validation {
	return Validation{
		Name: fmt.Sprintf("max(%v)", max),
		Check: func(value any) error {
			switch v := value.(type) {
			case int64:
				if float64(v) > max {
					return fmt.Errorf("value %d is greater than maximum %v", v, max)
				}
			case float64:
				if v > max {
					return fmt.Errorf("value %f is greater than maximum %v", v, max)
				}
			}
			return nil
		},
	}
}

func validateMinLength(min int) Validation {
	return Validation{
		Name: fmt.Sprintf("minLength(%d)", min),
		Check: func(value any) error {
			if s, ok := value.(string); ok {
				if len(s) < min {
					return fmt.Errorf("value length %d is less than minimum %d", len(s), min)
				}
			}
			return nil
		},
	}
}

func validateMaxLength(max int) Validation {
	return Validation{
		Name: fmt.Sprintf("maxLength(%d)", max),
		Check: func(value any) error {
			if s, ok := value.(string); ok {
				if len(s) > max {
					return fmt.Errorf("value length %d is greater than maximum %d", len(s), max)
				}
			}
			return nil
		},
	}
}

func validateRegexp(pattern string) Validation {
	return Validation{
		Name: fmt.Sprintf("regexp(%s)", pattern),
		Check: func(value any) error {
			if s, ok := value.(string); ok {
				re := validationCache.getRegexp(pattern)
				if !re.MatchString(s) {
					return fmt.Errorf("value does not match pattern %s", pattern)
				}
			}
			return nil
		},
	}
}

func validateOneOf(allowed []string) Validation {
	return Validation{
		Name: fmt.Sprintf("oneOf(%v)", allowed),
		Check: func(value any) error {
			if s, ok := value.(string); ok {
				for _, a := range allowed {
					if s == a {
						return nil
					}
				}
				return fmt.Errorf("value '%s' is not one of: %v", s, allowed)
			}
			return nil
		},
	}
}

func validateMinDuration(min time.Duration) Validation {
	return Validation{
		Name: fmt.Sprintf("minDuration(%s)", min),
		Check: func(value any) error {
			if d, ok := value.(time.Duration); ok {
				if d < min {
					return fmt.Errorf("duration %s is less than minimum %s", d, min)
				}
			}
			return nil
		},
	}
}

func validateMaxDuration(max time.Duration) Validation {
	return Validation{
		Name: fmt.Sprintf("maxDuration(%s)", max),
		Check: func(value any) error {
			if d, ok := value.(time.Duration); ok {
				if d > max {
					return fmt.Errorf("duration %s is greater than maximum %s", d, max)
				}
			}
			return nil
		},
	}
}

func validateMinItems(min int) Validation {
	return Validation{
		Name: fmt.Sprintf("minItems(%d)", min),
		Check: func(value any) error {
			switch v := value.(type) {
			case []string:
				if len(v) < min {
					return fmt.Errorf("array has %d items, minimum is %d", len(v), min)
				}
			case []int64:
				if len(v) < min {
					return fmt.Errorf("array has %d items, minimum is %d", len(v), min)
				}
			}
			return nil
		},
	}
}

func validateMaxItems(max int) Validation {
	return Validation{
		Name: fmt.Sprintf("maxItems(%d)", max),
		Check: func(value any) error {
			switch v := value.(type) {
			case []string:
				if len(v) > max {
					return fmt.Errorf("array has %d items, maximum is %d", len(v), max)
				}
			case []int64:
				if len(v) > max {
					return fmt.Errorf("array has %d items, maximum is %d", len(v), max)
				}
			}
			return nil
		},
	}
}

// FileMode validation methods
func validateValidFilePermission() Validation {
	return Validation{
		Name: "validFilePermission",
		Check: func(value any) error {
			if mode, ok := value.(os.FileMode); ok {
				// Check for common invalid permission bits (bits beyond standard rwx)
				// Standard permission bits are: 0o777 (user, group, other)
				// Common additional bits are: 0o4000 (suid), 0o2000 (sgid), 0o1000 (sticky)
				// So valid bits are typically: 0o7777
				if mode > 0o7777 {
					return fmt.Errorf("file mode %o contains invalid permission bits", mode)
				}
				return nil
			}
			return fmt.Errorf("value must be os.FileMode, got %T", value)
		},
	}
}

func validateFileModeRange(min, max os.FileMode) Validation {
	return Validation{
		Name: fmt.Sprintf("fileModeRange(%o, %o)", min, max),
		Check: func(value any) error {
			if mode, ok := value.(os.FileMode); ok {
				if mode < min || mode > max {
					return fmt.Errorf("file mode %o is not between %o and %o", mode, min, max)
				}
				return nil
			}
			return fmt.Errorf("value must be os.FileMode, got %T", value)
		},
	}
}

// Path validation methods
func validatePathExists() Validation {
	return Validation{
		Name: "pathExists",
		Check: func(value any) error {
			if path, ok := value.(string); ok {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					return fmt.Errorf("path '%s' does not exist", path)
				}
				return nil
			}
			return fmt.Errorf("value must be string, got %T", value)
		},
	}
}

func validatePathIsFile() Validation {
	return Validation{
		Name: "pathIsFile",
		Check: func(value any) error {
			if path, ok := value.(string); ok {
				info, err := os.Stat(path)
				if err != nil {
					return fmt.Errorf("cannot stat path '%s': %w", path, err)
				}
				if info.IsDir() {
					return fmt.Errorf("path '%s' is a directory, not a file", path)
				}
				return nil
			}
			return fmt.Errorf("value must be string, got %T", value)
		},
	}
}

func validatePathIsDir() Validation {
	return Validation{
		Name: "pathIsDir",
		Check: func(value any) error {
			if path, ok := value.(string); ok {
				info, err := os.Stat(path)
				if err != nil {
					return fmt.Errorf("cannot stat path '%s': %w", path, err)
				}
				if !info.IsDir() {
					return fmt.Errorf("path '%s' is a file, not a directory", path)
				}
				return nil
			}
			return fmt.Errorf("value must be string, got %T", value)
		},
	}
}

func validatePathReadable() Validation {
	return Validation{
		Name: "pathReadable",
		Check: func(value any) error {
			if path, ok := value.(string); ok {
				file, err := os.OpenFile(path, os.O_RDONLY, 0)
				if err != nil {
					return fmt.Errorf("path '%s' is not readable: %w", path, err)
				}
				file.Close()
				return nil
			}
			return fmt.Errorf("value must be string, got %T", value)
		},
	}
}

func validatePathWritable() Validation {
	return Validation{
		Name: "pathWritable",
		Check: func(value any) error {
			if path, ok := value.(string); ok {
				file, err := os.OpenFile(path, os.O_WRONLY, 0)
				if err != nil {
					return fmt.Errorf("path '%s' is not writable: %w", path, err)
				}
				file.Close()
				return nil
			}
			return fmt.Errorf("value must be string, got %T", value)
		},
	}
}

// IP validation methods
func validateIPVersion(version int) Validation {
	return Validation{
		Name: fmt.Sprintf("ipVersion(%d)", version),
		Check: func(value any) error {
			if ip, ok := value.(net.IP); ok {
				if version == 4 && ip.To4() == nil {
					return fmt.Errorf("IP '%s' is not IPv4", ip)
				}
				if version == 6 && ip.To4() != nil {
					return fmt.Errorf("IP '%s' is not IPv6", ip)
				}
				return nil
			}
			return fmt.Errorf("value must be net.IP, got %T", value)
		},
	}
}

func validateIPPrivate() Validation {
	return Validation{
		Name: "ipPrivate",
		Check: func(value any) error {
			if ip, ok := value.(net.IP); ok {
				if !ip.IsPrivate() {
					return fmt.Errorf("IP '%s' is not a private address", ip)
				}
				return nil
			}
			return fmt.Errorf("value must be net.IP, got %T", value)
		},
	}
}

func validateIPLoopback() Validation {
	return Validation{
		Name: "ipLoopback",
		Check: func(value any) error {
			if ip, ok := value.(net.IP); ok {
				if !ip.IsLoopback() {
					return fmt.Errorf("IP '%s' is not a loopback address", ip)
				}
				return nil
			}
			return fmt.Errorf("value must be net.IP, got %T", value)
		},
	}
}

// Time validation methods
func validateTimeAfter(after time.Time) Validation {
	return Validation{
		Name: fmt.Sprintf("timeAfter(%s)", after.Format(time.RFC3339)),
		Check: func(value any) error {
			if t, ok := value.(time.Time); ok {
				if !t.After(after) {
					return fmt.Errorf("time %s is not after %s", t.Format(time.RFC3339), after.Format(time.RFC3339))
				}
				return nil
			}
			return fmt.Errorf("value must be time.Time, got %T", value)
		},
	}
}

func validateTimeBefore(before time.Time) Validation {
	return Validation{
		Name: fmt.Sprintf("timeBefore(%s)", before.Format(time.RFC3339)),
		Check: func(value any) error {
			if t, ok := value.(time.Time); ok {
				if !t.Before(before) {
					return fmt.Errorf("time %s is not before %s", t.Format(time.RFC3339), before.Format(time.RFC3339))
				}
				return nil
			}
			return fmt.Errorf("value must be time.Time, got %T", value)
		},
	}
}

func validateTimeRange(min, max time.Time) Validation {
	return Validation{
		Name: fmt.Sprintf("timeRange(%s, %s)", min.Format(time.RFC3339), max.Format(time.RFC3339)),
		Check: func(value any) error {
			if t, ok := value.(time.Time); ok {
				if !t.After(min) || !t.Before(max) {
					return fmt.Errorf("time %s is not between %s and %s", t.Format(time.RFC3339), min.Format(time.RFC3339), max.Format(time.RFC3339))
				}
				return nil
			}
			return fmt.Errorf("value must be time.Time, got %T", value)
		},
	}
}
