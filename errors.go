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
	Message          string
	Display          string
	ErrorDescription string
}

func (e *ConfigError) Error() string {
	if e.ErrorDescription != "" {
		return e.ErrorDescription
	}
	if e.Source == "none" {
		return fmt.Sprintf("%s: %s", e.Key, e.Message)
	}
	if e.Value != "" {
		return fmt.Sprintf("%s (%s=%s): %s", e.Key, e.Source, e.Value, e.Message)
	}
	return fmt.Sprintf("%s (%s): %s", e.Key, e.Source, e.Message)
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

func standardizeValidationMessage(value any, validationName string, original error) string {
	switch {
	case validationName == "required":
		return "Not provided"
	case strings.HasPrefix(validationName, "min("):
		return fmt.Sprintf("Below minimum: %v", value)
	case strings.HasPrefix(validationName, "max("):
		return fmt.Sprintf("Out of bounds: %v", value)
	case strings.HasPrefix(validationName, "minLength("):
		return fmt.Sprintf("Too short: %q", value)
	case strings.HasPrefix(validationName, "maxLength("):
		return fmt.Sprintf("Too long: %q", value)
	case strings.HasPrefix(validationName, "regexp("):
		return fmt.Sprintf("Invalid format: %q", value)
	case strings.HasPrefix(validationName, "oneOf("):
		allowed := extractOneOfValues(validationName)
		return fmt.Sprintf("Invalid choice: %q (allowed: %s)", value, allowed)
	case strings.HasPrefix(validationName, "minDuration("):
		return fmt.Sprintf("Too short: %v", value)
	case strings.HasPrefix(validationName, "maxDuration("):
		return fmt.Sprintf("Too long: %v", value)
	case strings.HasPrefix(validationName, "minItems("):
		return fmt.Sprintf("Too few items: %v", value)
	case strings.HasPrefix(validationName, "maxItems("):
		return fmt.Sprintf("Too many items: %v", value)
	default:
		return original.Error()
	}
}

func newValidationConfigError(key string, def *Definition, source string, rawValue string, value any, validationName string, original error) ConfigError {
	return ConfigError{
		Key:              key,
		Source:           source,
		Value:            rawValue,
		Message:          original.Error(),
		Display:          buildErrorDisplay(def),
		ErrorDescription: standardizeValidationMessage(value, validationName, original),
	}
}

func newParseConfigError(key string, def *Definition, source string, rawValue string, original error) ConfigError {
	return ConfigError{
		Key:              key,
		Source:           source,
		Value:            rawValue,
		Message:          original.Error(),
		Display:          buildErrorDisplay(def),
		ErrorDescription: original.Error(),
	}
}

func newRequiredConfigError(key string, def *Definition) ConfigError {
	return ConfigError{
		Key:              key,
		Source:           "validation",
		Message:          "required value not provided",
		Display:          buildErrorDisplay(def),
		ErrorDescription: "Not provided",
	}
}

// formatErrors formats configuration errors for display
func formatErrors(errs []ConfigError) string {
	if len(errs) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("Configuration errors detected:\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	for i, err := range errs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("ERROR: %s\n", err.Key))
		if err.Source != "none" {
			sb.WriteString(fmt.Sprintf("  Source: %s\n", err.Source))
		}
		if err.Value != "" {
			sb.WriteString(fmt.Sprintf("  Value: %s\n", err.Value))
		}
		sb.WriteString(fmt.Sprintf("  Error: %s\n", err.Message))
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total: %d error(s)\n", len(errs)))

	return sb.String()
}

// maskSecret masks a secret value for display
func maskSecret(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
