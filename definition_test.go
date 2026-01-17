package commandkit

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDefinitionBuilderTypes(t *testing.T) {
	cfg := New()

	// Test all type setters
	cfg.Define("STRING_VAL").String()
	cfg.Define("INT64_VAL").Int64()
	cfg.Define("FLOAT64_VAL").Float64()
	cfg.Define("BOOL_VAL").Bool()
	cfg.Define("DURATION_VAL").Duration()
	cfg.Define("URL_VAL").URL()
	cfg.Define("STRING_SLICE_VAL").StringSlice()
	cfg.Define("INT64_SLICE_VAL").Int64Slice()

	// Verify types are set correctly
	if cfg.definitions["STRING_VAL"].valueType != TypeString {
		t.Error("String() should set TypeString")
	}
	if cfg.definitions["INT64_VAL"].valueType != TypeInt64 {
		t.Error("Int64() should set TypeInt64")
	}
	if cfg.definitions["FLOAT64_VAL"].valueType != TypeFloat64 {
		t.Error("Float64() should set TypeFloat64")
	}
	if cfg.definitions["BOOL_VAL"].valueType != TypeBool {
		t.Error("Bool() should set TypeBool")
	}
	if cfg.definitions["DURATION_VAL"].valueType != TypeDuration {
		t.Error("Duration() should set TypeDuration")
	}
	if cfg.definitions["URL_VAL"].valueType != TypeURL {
		t.Error("URL() should set TypeURL")
	}
	if cfg.definitions["STRING_SLICE_VAL"].valueType != TypeStringSlice {
		t.Error("StringSlice() should set TypeStringSlice")
	}
	if cfg.definitions["INT64_SLICE_VAL"].valueType != TypeInt64Slice {
		t.Error("Int64Slice() should set TypeInt64Slice")
	}
}

func TestDefinitionBuilderSources(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Env("PORT_ENV").Flag("port-flag")

	def := cfg.definitions["PORT"]
	if def.envVar != "PORT_ENV" {
		t.Errorf("Env() should set envVar, got %s", def.envVar)
	}
	if def.flag != "port-flag" {
		t.Errorf("Flag() should set flag, got %s", def.flag)
	}
}

func TestDefinitionBuilderBehaviors(t *testing.T) {
	cfg := New()

	cfg.Define("REQUIRED_VAL").String().Required()
	cfg.Define("SECRET_VAL").String().Secret()
	cfg.Define("DEFAULT_VAL").String().Default("default-value")
	cfg.Define("DELIM_VAL").StringSlice().Delimiter("|")
	cfg.Define("DESC_VAL").String().Description("Test description")

	if !cfg.definitions["REQUIRED_VAL"].required {
		t.Error("Required() should set required=true")
	}
	if !cfg.definitions["SECRET_VAL"].secret {
		t.Error("Secret() should set secret=true")
	}
	if cfg.definitions["DEFAULT_VAL"].defaultValue != "default-value" {
		t.Error("Default() should set defaultValue")
	}
	if cfg.definitions["DELIM_VAL"].delimiter != "|" {
		t.Error("Delimiter() should set delimiter")
	}
	if cfg.definitions["DESC_VAL"].description != "Test description" {
		t.Error("Description() should set description")
	}
}

func TestDefinitionBuilderNumericValidation(t *testing.T) {
	cfg := New()

	cfg.Define("MIN_VAL").Int64().Min(10)
	cfg.Define("MAX_VAL").Int64().Max(100)
	cfg.Define("RANGE_VAL").Int64().Range(1, 65535)

	// Min validation
	if len(cfg.definitions["MIN_VAL"].validations) != 1 {
		t.Error("Min() should add 1 validation")
	}

	// Max validation
	if len(cfg.definitions["MAX_VAL"].validations) != 1 {
		t.Error("Max() should add 1 validation")
	}

	// Range validation (adds 2: min and max)
	if len(cfg.definitions["RANGE_VAL"].validations) != 2 {
		t.Error("Range() should add 2 validations")
	}
}

func TestDefinitionBuilderStringValidation(t *testing.T) {
	cfg := New()

	cfg.Define("MIN_LEN").String().MinLength(5)
	cfg.Define("MAX_LEN").String().MaxLength(100)
	cfg.Define("LEN_RANGE").String().LengthRange(5, 100)
	cfg.Define("REGEXP").String().Regexp(`^[a-z]+$`)
	cfg.Define("ONE_OF").String().OneOf("a", "b", "c")

	if len(cfg.definitions["MIN_LEN"].validations) != 1 {
		t.Error("MinLength() should add 1 validation")
	}
	if len(cfg.definitions["MAX_LEN"].validations) != 1 {
		t.Error("MaxLength() should add 1 validation")
	}
	if len(cfg.definitions["LEN_RANGE"].validations) != 2 {
		t.Error("LengthRange() should add 2 validations")
	}
	if len(cfg.definitions["REGEXP"].validations) != 1 {
		t.Error("Regexp() should add 1 validation")
	}
	if len(cfg.definitions["ONE_OF"].validations) != 1 {
		t.Error("OneOf() should add 1 validation")
	}
}

func TestDefinitionBuilderDurationValidation(t *testing.T) {
	cfg := New()

	cfg.Define("MIN_DUR").Duration().MinDuration(5 * time.Second)
	cfg.Define("MAX_DUR").Duration().MaxDuration(1 * time.Minute)
	cfg.Define("DUR_RANGE").Duration().DurationRange(5*time.Second, 1*time.Minute)

	if len(cfg.definitions["MIN_DUR"].validations) != 1 {
		t.Error("MinDuration() should add 1 validation")
	}
	if len(cfg.definitions["MAX_DUR"].validations) != 1 {
		t.Error("MaxDuration() should add 1 validation")
	}
	if len(cfg.definitions["DUR_RANGE"].validations) != 2 {
		t.Error("DurationRange() should add 2 validations")
	}
}

func TestDefinitionBuilderDurationSecValidation(t *testing.T) {
	cfg := New()

	cfg.Define("MIN_DUR_SEC").Duration().MinDurationSec(5)
	cfg.Define("MAX_DUR_SEC").Duration().MaxDurationSec(60)
	cfg.Define("DUR_RANGE_SEC").Duration().DurationRangeSec(5, 60)

	if len(cfg.definitions["MIN_DUR_SEC"].validations) != 1 {
		t.Error("MinDurationSec() should add 1 validation")
	}
	if len(cfg.definitions["MAX_DUR_SEC"].validations) != 1 {
		t.Error("MaxDurationSec() should add 1 validation")
	}
	if len(cfg.definitions["DUR_RANGE_SEC"].validations) != 2 {
		t.Error("DurationRangeSec() should add 2 validations")
	}

	// Test that the validations work correctly
	os.Setenv("MIN_DUR_SEC", "10s")
	defer os.Unsetenv("MIN_DUR_SEC")

	cfg.definitions["MIN_DUR_SEC"].envVar = "MIN_DUR_SEC"
	errs := cfg.Process()
	if len(errs) > 0 {
		t.Errorf("Unexpected errors: %v", errs)
	}
}

func TestDefinitionBuilderArrayValidation(t *testing.T) {
	cfg := New()

	cfg.Define("MIN_ITEMS").StringSlice().MinItems(2)
	cfg.Define("MAX_ITEMS").StringSlice().MaxItems(5)
	cfg.Define("ITEMS_RANGE").StringSlice().ItemsRange(2, 5)

	if len(cfg.definitions["MIN_ITEMS"].validations) != 1 {
		t.Error("MinItems() should add 1 validation")
	}
	if len(cfg.definitions["MAX_ITEMS"].validations) != 1 {
		t.Error("MaxItems() should add 1 validation")
	}
	if len(cfg.definitions["ITEMS_RANGE"].validations) != 2 {
		t.Error("ItemsRange() should add 2 validations")
	}
}

func TestDefinitionBuilderCustomValidation(t *testing.T) {
	cfg := New()

	evenLength := func(value any) error {
		if s, ok := value.(string); ok && len(s)%2 != 0 {
			return fmt.Errorf("string length must be even")
		}
		return nil
	}

	cfg.Define("CUSTOM").String().Custom("even-length", evenLength).Default("ab")

	if len(cfg.definitions["CUSTOM"].validations) != 1 {
		t.Error("Custom() should add 1 validation")
	}

	// Test that custom validation works
	errs := cfg.Process()
	if len(errs) > 0 {
		t.Errorf("Unexpected errors for valid value: %v", errs)
	}

	// Test with invalid value
	cfg2 := New()
	cfg2.Define("CUSTOM").String().Custom("even-length", evenLength).Default("abc")
	errs = cfg2.Process()
	if len(errs) != 1 {
		t.Errorf("Expected 1 error for odd-length string, got %d", len(errs))
	}
}

func TestDefinitionBuilderChaining(t *testing.T) {
	cfg := New()

	// Test that all methods can be chained
	cfg.Define("FULL").
		String().
		Env("FULL_ENV").
		Flag("full-flag").
		Default("default").
		Required().
		MinLength(3).
		MaxLength(100).
		Description("Full test")

	def := cfg.definitions["FULL"]
	if def.valueType != TypeString {
		t.Error("Chaining: type not set")
	}
	if def.envVar != "FULL_ENV" {
		t.Error("Chaining: envVar not set")
	}
	if def.flag != "full-flag" {
		t.Error("Chaining: flag not set")
	}
	if def.defaultValue != "default" {
		t.Error("Chaining: defaultValue not set")
	}
	if !def.required {
		t.Error("Chaining: required not set")
	}
	if def.description != "Full test" {
		t.Error("Chaining: description not set")
	}
	if len(def.validations) != 3 { // required + minLength + maxLength
		t.Errorf("Chaining: expected 3 validations, got %d", len(def.validations))
	}
}

func TestDefinitionBuilderBuild(t *testing.T) {
	cfg := New()

	builder := newDefinitionBuilder(cfg, "TEST")
	builder.String().Default("test")

	def := builder.build()
	if def.key != "TEST" {
		t.Error("build() should return definition with correct key")
	}
	if def.defaultValue != "test" {
		t.Error("build() should return definition with correct default")
	}
}
