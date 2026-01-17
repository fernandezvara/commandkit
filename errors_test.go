package commandkit

import (
	"strings"
	"testing"
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
