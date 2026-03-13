// commandkit/errors.go
package commandkit

import (
	"fmt"
	"io/fs"
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
	} else if def.envVar != "" {
		// Use unified display for environment-only variables
		return buildDefinitionDisplay(def)
	} else {
		base = fmt.Sprintf("%s %s", def.key, valueType)
	}

	if len(indicators) == 0 {
		return base
	}

	for _, indicator := range indicators {
		base += fmt.Sprintf(" (%s)", indicator)
	}
	return base
}

// buildDefinitionDisplay creates unified display for both flags and environment variables
func buildDefinitionDisplay(def *Definition) string {
	valueType := def.valueType.String()
	var indicators []string

	// Collect all indicators
	if shouldDisplayDefault(def) {
		var defaultDisplay string
		if def.valueType == TypeFileMode {
			// Special handling for FileMode - display in octal format with leading zero
			if mode, ok := def.defaultValue.(fs.FileMode); ok {
				defaultDisplay = fmt.Sprintf("default: %#o", mode)
			} else if intValue, ok := def.defaultValue.(int); ok {
				// Handle case where FileMode is stored as int (e.g., from 0640 literal)
				defaultDisplay = fmt.Sprintf("default: %#o", intValue)
			} else if intValue64, ok := def.defaultValue.(int64); ok {
				// Handle case where FileMode is stored as int64
				defaultDisplay = fmt.Sprintf("default: %#o", intValue64)
			} else {
				defaultDisplay = fmt.Sprintf("default: %v", def.defaultValue)
			}
		} else {
			defaultDisplay = fmt.Sprintf("default: %v", def.defaultValue)
		}
		indicators = append(indicators, defaultDisplay)
	}
	if def.required {
		indicators = append(indicators, "required")
	}
	if def.secret {
		indicators = append(indicators, "secret")
	}

	// Add validations
	validations := formatValidation(def.validations)
	indicators = append(indicators, validations...)

	// Determine the base format based on what type of display this is
	var base string
	if def.flag != "" {
		// This is a flag display
		base = fmt.Sprintf("--%s %s", def.flag, valueType)

		// Add env indicator if it also has an environment variable
		if def.envVar != "" {
			indicators = append(indicators, fmt.Sprintf("env: %s", def.envVar))
		}
	} else if def.envVar != "" {
		// This is an environment-only variable display
		base = fmt.Sprintf("%s %s", def.envVar, valueType)
		// Note: Don't add "env: VARNAME" since this IS the env var display
	} else {
		// No flag or env var (fallback case)
		base = fmt.Sprintf("(no flag) %s", valueType)
	}

	// Return with indicators if any
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
