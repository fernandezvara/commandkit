package commandkit

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestConfigErrorString(t *testing.T) {
	tests := []struct {
		err      ConfigError
		expected string
	}{
		{
			ConfigError{Key: "PORT", Source: "env", Value: "8080", Message: "invalid value"},
			"PORT (env=8080): invalid value",
		},
		{
			ConfigError{Key: "HOST", Source: "flag", Value: "", Message: "required"},
			"HOST (flag): required",
		},
		{
			ConfigError{Key: "API_KEY", Source: "none", Value: "", Message: "not provided"},
			"API_KEY: not provided",
		},
	}

	for _, tt := range tests {
		result := tt.err.Error()
		if result != tt.expected {
			t.Errorf("ConfigError.Error() = %q, expected %q", result, tt.expected)
		}
	}
}

func TestShouldDisplayDefault(t *testing.T) {
	tests := []struct {
		name     string
		def      *Definition
		expected bool
	}{
		{
			name:     "nil default",
			def:      &Definition{valueType: TypeString, defaultValue: nil},
			expected: false,
		},
		{
			name:     "bool false default hidden",
			def:      &Definition{valueType: TypeBool, defaultValue: false},
			expected: false,
		},
		{
			name:     "bool true default shown",
			def:      &Definition{valueType: TypeBool, defaultValue: true},
			expected: true,
		},
		{
			name:     "string default shown",
			def:      &Definition{valueType: TypeString, defaultValue: "info"},
			expected: true,
		},
	}

	for _, tt := range tests {
		if got := shouldDisplayDefault(tt.def); got != tt.expected {
			t.Fatalf("%s: expected %v, got %v", tt.name, tt.expected, got)
		}
	}
}

func TestCleanValidationDisplay(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "oneOf: ['debug', 'info', 'warn', 'error']",
			expected: "oneOf: debug info warn error",
		},
		{
			input:    "valid: 1-65535",
			expected: "valid: 1-65535",
		},
	}

	for _, tt := range tests {
		if got := cleanValidationDisplay(tt.input); got != tt.expected {
			t.Fatalf("cleanValidationDisplay(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildErrorDisplay(t *testing.T) {
	flagDef := &Definition{
		key:          "PORT",
		valueType:    TypeInt64,
		flag:         "port",
		defaultValue: int64(8080),
		required:     true,
	}

	if got := buildErrorDisplay(flagDef); got != "--port int64 (required) (default: 8080)" {
		t.Fatalf("unexpected flag error display: %q", got)
	}

	envDef := &Definition{
		key:       "DATABASE_URL",
		valueType: TypeString,
		envVar:    "DATABASE_URL",
		required:  true,
	}

	if got := buildErrorDisplay(envDef); got != "(no flag) string (env: DATABASE_URL, required)" {
		t.Fatalf("unexpected env-only error display: %q", got)
	}
}

func TestBuildFlagDisplay(t *testing.T) {
	def := &Definition{
		key:          "LOG_LEVEL",
		valueType:    TypeString,
		flag:         "log-level",
		envVar:       "LOG_LEVEL",
		defaultValue: "info",
		validations: []Validation{
			{Name: "oneOf(debug,info,warn,error)"},
		},
	}

	got := buildFlagDisplay(def)
	expectedParts := []string{
		"--log-level string",
		"(default: info)",
		"(oneOf: debug info warn error)",
		"(env: LOG_LEVEL)",
	}

	for _, part := range expectedParts {
		if !strings.Contains(got, part) {
			t.Fatalf("expected %q in %q", part, got)
		}
	}
}

func TestStandardizeValidationMessage(t *testing.T) {
	original := fmt.Errorf("original error")

	tests := []struct {
		name           string
		value          any
		validationName string
		expected       string
	}{
		{"required", nil, "required", "Not provided"},
		{"min", int64(1), "min(10)", "Below minimum: 1"},
		{"max", int64(80000), "max(65535)", "Out of bounds: 80000"},
		{"minLength", "ab", "minLength(3)", `Too short: "ab"`},
		{"maxLength", "abcdef", "maxLength(3)", `Too long: "abcdef"`},
		{"regexp", "abc", "regexp(^\\d+$)", `Invalid format: "abc"`},
		{"oneOf", "s", "oneOf(debug,info,warn,error)", `Invalid choice: "s" (allowed: ['debug', 'info', 'warn', 'error'])`},
		{"minDuration", 5 * time.Second, "minDuration(10s)", "Too short: 5s"},
		{"maxDuration", 2 * time.Hour, "maxDuration(1h)", "Too long: 2h0m0s"},
		{"minItems", []string{"a"}, "minItems(2)", "Too few items: [a]"},
		{"maxItems", []string{"a", "b", "c"}, "maxItems(2)", "Too many items: [a b c]"},
		{"default", "x", "custom", "original error"},
	}

	for _, tt := range tests {
		got := standardizeValidationMessage(tt.value, tt.validationName, original)
		if got != tt.expected {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.expected, got)
		}
	}
}

func TestNewConfigErrorConstructors(t *testing.T) {
	def := &Definition{
		key:          "PORT",
		valueType:    TypeInt64,
		flag:         "port",
		defaultValue: int64(8080),
	}

	validationErr := newValidationConfigError("PORT", def, "flag", "80000", int64(80000), "max(65535)", fmt.Errorf("value 80000 is greater than maximum 65535"))
	if validationErr.Display != "--port int64 (default: 8080)" {
		t.Fatalf("unexpected validation display: %q", validationErr.Display)
	}
	if validationErr.ErrorDescription != "Out of bounds: 80000" {
		t.Fatalf("unexpected validation error description: %q", validationErr.ErrorDescription)
	}

	parseErr := newParseConfigError("PORT", def, "flag", "abc", fmt.Errorf("invalid syntax"))
	if parseErr.ErrorDescription != "invalid syntax" {
		t.Fatalf("unexpected parse error description: %q", parseErr.ErrorDescription)
	}

	requiredDef := &Definition{
		key:       "BASE_URL",
		valueType: TypeURL,
		flag:      "base-url",
		required:  true,
	}
	requiredErr := newRequiredConfigError("BASE_URL", requiredDef)
	if requiredErr.Display != "--base-url url (required)" {
		t.Fatalf("unexpected required display: %q", requiredErr.Display)
	}
	if requiredErr.ErrorDescription != "Not provided" {
		t.Fatalf("unexpected required error description: %q", requiredErr.ErrorDescription)
	}
}

func TestFormatErrors(t *testing.T) {
	// Test empty errors
	result := formatErrors(nil)
	if result != "" {
		t.Error("formatErrors(nil) should return empty string")
	}

	result = formatErrors([]ConfigError{})
	if result != "" {
		t.Error("formatErrors([]) should return empty string")
	}

	// Test single error
	errs := []ConfigError{
		{Key: "PORT", Source: "env", Value: "invalid", Message: "must be a number"},
	}
	result = formatErrors(errs)

	if !strings.Contains(result, "Configuration errors detected:") {
		t.Error("formatErrors should contain header")
	}
	if !strings.Contains(result, "PORT") {
		t.Error("formatErrors should contain key name")
	}
	if !strings.Contains(result, "must be a number") {
		t.Error("formatErrors should contain error message")
	}
	if !strings.Contains(result, "Total: 1 error(s)") {
		t.Error("formatErrors should contain error count")
	}

	// Test multiple errors
	errs = []ConfigError{
		{Key: "PORT", Source: "env", Value: "99999", Message: "value out of range"},
		{Key: "API_KEY", Source: "none", Value: "", Message: "required value not provided"},
		{Key: "HOST", Source: "flag", Value: "invalid-url", Message: "invalid URL format"},
	}
	result = formatErrors(errs)

	if !strings.Contains(result, "Total: 3 error(s)") {
		t.Error("formatErrors should show correct error count")
	}
	if !strings.Contains(result, "PORT") {
		t.Error("formatErrors should contain PORT")
	}
	if !strings.Contains(result, "API_KEY") {
		t.Error("formatErrors should contain API_KEY")
	}
	if !strings.Contains(result, "HOST") {
		t.Error("formatErrors should contain HOST")
	}

	// Check simple formatting
	if !strings.Contains(result, "Configuration errors detected:") {
		t.Error("formatErrors should use simple formatting")
	}
}

func TestFormatErrorsSourceDisplay(t *testing.T) {
	// Test that source "none" doesn't show source line
	errs := []ConfigError{
		{Key: "API_KEY", Source: "none", Value: "", Message: "required"},
	}
	result := formatErrors(errs)

	// Should not contain "Source: none"
	if strings.Contains(result, "Source: none") {
		t.Error("formatErrors should not show 'Source: none'")
	}

	// Test that other sources show source line
	errs = []ConfigError{
		{Key: "PORT", Source: "env", Value: "8080", Message: "invalid"},
	}
	result = formatErrors(errs)

	if !strings.Contains(result, "Source: env") {
		t.Error("formatErrors should show 'Source: env'")
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "****"},
		{"a", "****"},
		{"ab", "****"},
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "ab*de"},
		{"abcdef", "ab**ef"},
		{"secret-key-12345", "se************45"},
		{"my-super-secret-api-key", "my*******************ey"},
	}

	for _, tt := range tests {
		result := maskSecret(tt.input)
		if result != tt.expected {
			t.Errorf("maskSecret(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestMaskSecretLength(t *testing.T) {
	// For strings longer than 4 chars, masked length should equal original length
	inputs := []string{"hello", "password123", "super-secret-key"}

	for _, input := range inputs {
		result := maskSecret(input)
		if len(result) != len(input) {
			t.Errorf("maskSecret(%q) length = %d, expected %d", input, len(result), len(input))
		}
	}
}
