// commandkit/errors.go
package commandkit

import (
	"fmt"
	"strings"
)

func shouldDisplayDefault(def *Definition) bool {
	if def.defaultValue == nil {
		return false
	}
	if def.valueType == TypeBool {
		if value, ok := def.defaultValue.(bool); ok && !value {
			return false
		}
	}
	return true
}

func cleanValidationDisplay(validation string) string {
	if strings.HasPrefix(validation, "oneOf: [") && strings.HasSuffix(validation, "]") {
		value := strings.TrimPrefix(validation, "oneOf: [")
		value = strings.TrimSuffix(value, "]")
		value = strings.ReplaceAll(value, "'", "")
		value = strings.ReplaceAll(value, ",", "")
		value = strings.Join(strings.Fields(value), " ")
		return "oneOf: " + value
	}
	return validation
}

// ConfigError represents a single configuration error
type ConfigError struct {
	Key              string
	Source           string // "env", "flag", "default", or "none"
	Value            string // Masked if secret
	Display          string
	ErrorDescription string
}

func (e *ConfigError) Error() string {
	return e.ErrorDescription
}

func buildErrorDisplay(def *Definition) string {
	valueType := def.valueType.String()
	var indicators []string

	if def.envVar != "" {
		indicators = append(indicators, fmt.Sprintf("env: %s", def.envVar))
	}
	if def.required {
		indicators = append(indicators, "required")
	}
	if shouldDisplayDefault(def) {
		indicators = append(indicators, fmt.Sprintf("default: %v", def.defaultValue))
	}

	var base string
	if def.flag != "" {
		base = fmt.Sprintf("--%s %s", def.flag, valueType)
	} else {
		base = fmt.Sprintf("(no flag) %s", valueType)
	}

	if len(indicators) == 0 {
		return base
	}

	if def.flag != "" {
		for _, indicator := range indicators {
			base += fmt.Sprintf(" (%s)", indicator)
		}
		return base
	}

	return fmt.Sprintf("%s (%s)", base, strings.Join(indicators, ", "))
}

func buildFlagDisplay(def *Definition) string {
	valueType := def.valueType.String()
	var indicators []string

	if def.envVar != "" {
		indicators = append(indicators, fmt.Sprintf("env: %s", def.envVar))
	}
	if def.required {
		indicators = append(indicators, "required")
	}
	if shouldDisplayDefault(def) {
		indicators = append(indicators, fmt.Sprintf("default: %v", def.defaultValue))
	}
	validations := formatValidation(def.validations)
	if def.flag != "" {
		// For flags, emit each indicator as its own group to match help output.
		var base string
		base = fmt.Sprintf("--%s %s", def.flag, valueType)
		if def.required {
			base += " (required)"
		}
		if shouldDisplayDefault(def) {
			base += fmt.Sprintf(" (default: %v)", def.defaultValue)
		}
		for _, validation := range validations {
			base += fmt.Sprintf(" (%s)", cleanValidationDisplay(validation))
		}
		if def.envVar != "" {
			base += fmt.Sprintf(" (env: %s)", def.envVar)
		}
		if def.secret {
			base += " (secret)"
		}
		return base
	}

	indicators = append(indicators, validations...)
	if def.secret {
		indicators = append(indicators, "secret")
	}

	base := fmt.Sprintf("(no flag) %s", valueType)

	if len(indicators) == 0 {
		return base
	}

	return fmt.Sprintf("%s (%s)", base, strings.Join(indicators, ", "))
}

// newConfigError creates a unified ConfigError for all error types
func newConfigError(key string, def *Definition, source string, rawValue string, original error) ConfigError {
	return ConfigError{
		Key:              key,
		Source:           source,
		Value:            rawValue,
		Display:          buildErrorDisplay(def),
		ErrorDescription: original.Error(),
	}
}

// maskSecret masks a secret value for display
func maskSecret(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
