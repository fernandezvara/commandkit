// commandkit/definition.go
package commandkit

import (
	"fmt"
	"strings"
	"time"
)

// Definition represents a configuration definition with fluent builder
type Definition struct {
	key          string
	valueType    ValueType
	envVar       string
	flag         string
	defaultValue any
	required     bool
	secret       bool
	delimiter    string
	validations  []Validation
	description  string
}

// DefinitionBuilder provides a fluent API for building definitions
type DefinitionBuilder struct {
	def    *Definition
	config *Config
}

// formatValidation formats validation rules for help display
func formatValidation(validations []Validation) []string {
	var result []string
	var minVal, maxVal string

	for _, validation := range validations {
		switch {
		case validation.Name == "required":
			// Skip required as it's handled separately
			continue
		case strings.HasPrefix(validation.Name, "min("):
			minVal = extractValue(validation.Name, "min(")
		case strings.HasPrefix(validation.Name, "max("):
			maxVal = extractValue(validation.Name, "max(")
		case strings.HasPrefix(validation.Name, "oneOf("):
			// Extract values from oneOf(format)
			values := extractOneOfValues(validation.Name)
			result = append(result, fmt.Sprintf("oneOf: %s", values))
		case strings.HasPrefix(validation.Name, "minLength("):
			min := extractValue(validation.Name, "minLength(")
			result = append(result, fmt.Sprintf("minLength: %s", min))
		case strings.HasPrefix(validation.Name, "maxLength("):
			max := extractValue(validation.Name, "maxLength(")
			result = append(result, fmt.Sprintf("maxLength: %s", max))
		case strings.HasPrefix(validation.Name, "regexp("):
			pattern := extractValue(validation.Name, "regexp(")
			result = append(result, fmt.Sprintf("pattern: %s", pattern))
		default:
			// For other validations, use the name as-is
			result = append(result, validation.Name)
		}
	}

	// Handle min/max range
	if minVal != "" && maxVal != "" {
		result = append([]string{fmt.Sprintf("valid: %s-%s", minVal, maxVal)}, result...)
	} else if minVal != "" {
		result = append([]string{fmt.Sprintf("min: %s", minVal)}, result...)
	} else if maxVal != "" {
		result = append([]string{fmt.Sprintf("max: %s", maxVal)}, result...)
	}

	return result
}

// extractValue extracts numeric value from validation name like "min(8080)"
func extractValue(name, prefix string) string {
	start := strings.Index(name, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	end := strings.Index(name[start:], ")")
	if end == -1 {
		return name[start:]
	}
	return name[start : start+end]
}

// extractOneOfValues extracts values from oneOf(['a', 'b', 'c']) format
func extractOneOfValues(name string) string {
	start := strings.Index(name, "oneOf(")
	if start == -1 {
		return ""
	}
	start += len("oneOf(")
	end := strings.Index(name[start:], ")")
	if end == -1 {
		return name[start:]
	}
	values := name[start : start+end]

	// Handle array format oneOf([debug info warn error])
	if strings.HasPrefix(values, "[") && strings.HasSuffix(values, "]") {
		// Remove brackets and clean up
		content := values[1 : len(values)-1]
		// Split by space and filter empty strings
		parts := strings.Fields(content)
		var quotedParts []string
		for _, part := range parts {
			if part != "" {
				quotedParts = append(quotedParts, fmt.Sprintf("'%s'", part))
			}
		}
		return "[" + strings.Join(quotedParts, ", ") + "]"
	}

	// Handle simple format oneOf(a,b,c)
	parts := strings.Split(values, ",")
	var quotedParts []string
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			quotedParts = append(quotedParts, fmt.Sprintf("'%s'", strings.TrimSpace(part)))
		}
	}
	return "[" + strings.Join(quotedParts, ", ") + "]"
}

// formatFlagHelp generates enhanced help text with required/default indicators and validations
func formatFlagHelp(def *Definition) string {
	var indicators []string

	// 1. Environment variable context
	if def.envVar != "" {
		indicators = append(indicators, fmt.Sprintf("env: %s", def.envVar))
	}

	// 2. Required indicator
	if def.required {
		indicators = append(indicators, "required")
	}

	// 3. Default value (masked for secrets)
	if def.defaultValue != nil {
		if def.secret {
			indicators = append(indicators, "default: '[hidden]'")
		} else if def.valueType == TypeString {
			indicators = append(indicators, fmt.Sprintf("default: '%v'", def.defaultValue))
		} else {
			indicators = append(indicators, fmt.Sprintf("default: %v", def.defaultValue))
		}
	}

	// 4. Validations
	validations := formatValidation(def.validations)
	indicators = append(indicators, validations...)

	// 5. Secret indicator
	if def.secret {
		indicators = append(indicators, "secret")
	}

	// Combine description with indicators
	if len(indicators) > 0 {
		return fmt.Sprintf("%s (%s)", def.description, strings.Join(indicators, ", "))
	}

	return def.description
}

// newDefinitionBuilder creates a new builder
func newDefinitionBuilder(cfg *Config, key string) *DefinitionBuilder {
	return &DefinitionBuilder{
		def: &Definition{
			key:       key,
			valueType: TypeString, // default
			delimiter: ",",        // default delimiter
		},
		config: cfg,
	}
}

// Type setters

func (b *DefinitionBuilder) String() *DefinitionBuilder {
	b.def.valueType = TypeString
	return b
}

func (b *DefinitionBuilder) Int64() *DefinitionBuilder {
	b.def.valueType = TypeInt64
	return b
}

func (b *DefinitionBuilder) Float64() *DefinitionBuilder {
	b.def.valueType = TypeFloat64
	return b
}

func (b *DefinitionBuilder) Bool() *DefinitionBuilder {
	b.def.valueType = TypeBool
	return b
}

func (b *DefinitionBuilder) Duration() *DefinitionBuilder {
	b.def.valueType = TypeDuration
	return b
}

func (b *DefinitionBuilder) URL() *DefinitionBuilder {
	b.def.valueType = TypeURL
	return b
}

func (b *DefinitionBuilder) StringSlice() *DefinitionBuilder {
	b.def.valueType = TypeStringSlice
	return b
}

func (b *DefinitionBuilder) Int64Slice() *DefinitionBuilder {
	b.def.valueType = TypeInt64Slice
	return b
}

// Source setters

func (b *DefinitionBuilder) Env(envVar string) *DefinitionBuilder {
	b.def.envVar = envVar
	return b
}

func (b *DefinitionBuilder) Flag(flag string) *DefinitionBuilder {
	b.def.flag = flag
	return b
}

// Behavior setters

func (b *DefinitionBuilder) Required() *DefinitionBuilder {
	b.def.required = true
	b.def.validations = append(b.def.validations, validateRequired())
	return b
}

func (b *DefinitionBuilder) Secret() *DefinitionBuilder {
	b.def.secret = true
	return b
}

func (b *DefinitionBuilder) Default(value any) *DefinitionBuilder {
	b.def.defaultValue = value
	return b
}

func (b *DefinitionBuilder) Delimiter(d string) *DefinitionBuilder {
	b.def.delimiter = d
	return b
}

func (b *DefinitionBuilder) Description(desc string) *DefinitionBuilder {
	b.def.description = desc
	return b
}

// Validation setters

func (b *DefinitionBuilder) Min(min float64) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMin(min))
	return b
}

func (b *DefinitionBuilder) Max(max float64) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMax(max))
	return b
}

func (b *DefinitionBuilder) Range(min, max float64) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMin(min))
	b.def.validations = append(b.def.validations, validateMax(max))
	return b
}

func (b *DefinitionBuilder) MinLength(min int) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMinLength(min))
	return b
}

func (b *DefinitionBuilder) MaxLength(max int) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMaxLength(max))
	return b
}

func (b *DefinitionBuilder) LengthRange(min, max int) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMinLength(min))
	b.def.validations = append(b.def.validations, validateMaxLength(max))
	return b
}

func (b *DefinitionBuilder) Regexp(pattern string) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateRegexp(pattern))
	return b
}

func (b *DefinitionBuilder) OneOf(allowed ...string) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateOneOf(allowed))
	return b
}

func (b *DefinitionBuilder) MinDuration(min time.Duration) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMinDuration(min))
	return b
}

func (b *DefinitionBuilder) MaxDuration(max time.Duration) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMaxDuration(max))
	return b
}

func (b *DefinitionBuilder) DurationRange(min, max time.Duration) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMinDuration(min))
	b.def.validations = append(b.def.validations, validateMaxDuration(max))
	return b
}

// Duration-specific validation methods that work with time.Duration
func (b *DefinitionBuilder) MinDurationSec(minSeconds float64) *DefinitionBuilder {
	min := time.Duration(minSeconds * float64(time.Second))
	b.def.validations = append(b.def.validations, validateMinDuration(min))
	return b
}

func (b *DefinitionBuilder) MaxDurationSec(maxSeconds float64) *DefinitionBuilder {
	max := time.Duration(maxSeconds * float64(time.Second))
	b.def.validations = append(b.def.validations, validateMaxDuration(max))
	return b
}

func (b *DefinitionBuilder) DurationRangeSec(minSeconds, maxSeconds float64) *DefinitionBuilder {
	min := time.Duration(minSeconds * float64(time.Second))
	max := time.Duration(maxSeconds * float64(time.Second))
	b.def.validations = append(b.def.validations, validateMinDuration(min))
	b.def.validations = append(b.def.validations, validateMaxDuration(max))
	return b
}

func (b *DefinitionBuilder) MinItems(min int) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMinItems(min))
	return b
}

func (b *DefinitionBuilder) MaxItems(max int) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMaxItems(max))
	return b
}

func (b *DefinitionBuilder) ItemsRange(min, max int) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, validateMinItems(min))
	b.def.validations = append(b.def.validations, validateMaxItems(max))
	return b
}

// Custom adds a custom validation function
func (b *DefinitionBuilder) Custom(name string, check func(value any) error) *DefinitionBuilder {
	b.def.validations = append(b.def.validations, Validation{Name: name, Check: check})
	return b
}

// Build finalizes the definition and adds it to the config
// This is called automatically; you don't need to call it explicitly
func (b *DefinitionBuilder) build() *Definition {
	return b.def
}
