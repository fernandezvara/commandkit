// commandkit/validation.go
package commandkit

import (
	"fmt"
	"regexp"
	"time"
)

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
	re := regexp.MustCompile(pattern)
	return Validation{
		Name: fmt.Sprintf("regexp(%s)", pattern),
		Check: func(value any) error {
			if s, ok := value.(string); ok {
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
