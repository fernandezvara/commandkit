// commandkit/definition.go
package commandkit

import (
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
