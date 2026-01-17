package commandkit

import (
	"testing"
	"time"
)

func TestValidateRequired(t *testing.T) {
	v := validateRequired()

	// Should fail for nil
	if err := v.Check(nil); err == nil {
		t.Error("validateRequired should fail for nil")
	}

	// Should fail for empty string
	if err := v.Check(""); err == nil {
		t.Error("validateRequired should fail for empty string")
	}

	// Should pass for non-empty values
	if err := v.Check("hello"); err != nil {
		t.Errorf("validateRequired should pass for non-empty string: %v", err)
	}

	if err := v.Check(123); err != nil {
		t.Errorf("validateRequired should pass for int: %v", err)
	}
}

func TestValidateMin(t *testing.T) {
	v := validateMin(10)

	// Int64 tests
	if err := v.Check(int64(5)); err == nil {
		t.Error("validateMin(10) should fail for int64(5)")
	}
	if err := v.Check(int64(10)); err != nil {
		t.Errorf("validateMin(10) should pass for int64(10): %v", err)
	}
	if err := v.Check(int64(15)); err != nil {
		t.Errorf("validateMin(10) should pass for int64(15): %v", err)
	}

	// Float64 tests
	if err := v.Check(float64(5.5)); err == nil {
		t.Error("validateMin(10) should fail for float64(5.5)")
	}
	if err := v.Check(float64(10.0)); err != nil {
		t.Errorf("validateMin(10) should pass for float64(10.0): %v", err)
	}
	if err := v.Check(float64(15.5)); err != nil {
		t.Errorf("validateMin(10) should pass for float64(15.5): %v", err)
	}

	// Non-numeric should pass (no validation)
	if err := v.Check("string"); err != nil {
		t.Errorf("validateMin should pass for non-numeric types: %v", err)
	}
}

func TestValidateMax(t *testing.T) {
	v := validateMax(100)

	// Int64 tests
	if err := v.Check(int64(150)); err == nil {
		t.Error("validateMax(100) should fail for int64(150)")
	}
	if err := v.Check(int64(100)); err != nil {
		t.Errorf("validateMax(100) should pass for int64(100): %v", err)
	}
	if err := v.Check(int64(50)); err != nil {
		t.Errorf("validateMax(100) should pass for int64(50): %v", err)
	}

	// Float64 tests
	if err := v.Check(float64(150.5)); err == nil {
		t.Error("validateMax(100) should fail for float64(150.5)")
	}
	if err := v.Check(float64(100.0)); err != nil {
		t.Errorf("validateMax(100) should pass for float64(100.0): %v", err)
	}
	if err := v.Check(float64(50.5)); err != nil {
		t.Errorf("validateMax(100) should pass for float64(50.5): %v", err)
	}
}

func TestValidateMinLength(t *testing.T) {
	v := validateMinLength(5)

	if err := v.Check("abc"); err == nil {
		t.Error("validateMinLength(5) should fail for 'abc'")
	}
	if err := v.Check("abcde"); err != nil {
		t.Errorf("validateMinLength(5) should pass for 'abcde': %v", err)
	}
	if err := v.Check("abcdefgh"); err != nil {
		t.Errorf("validateMinLength(5) should pass for 'abcdefgh': %v", err)
	}

	// Non-string should pass
	if err := v.Check(123); err != nil {
		t.Errorf("validateMinLength should pass for non-string: %v", err)
	}
}

func TestValidateMaxLength(t *testing.T) {
	v := validateMaxLength(5)

	if err := v.Check("abcdefgh"); err == nil {
		t.Error("validateMaxLength(5) should fail for 'abcdefgh'")
	}
	if err := v.Check("abcde"); err != nil {
		t.Errorf("validateMaxLength(5) should pass for 'abcde': %v", err)
	}
	if err := v.Check("abc"); err != nil {
		t.Errorf("validateMaxLength(5) should pass for 'abc': %v", err)
	}
}

func TestValidateRegexp(t *testing.T) {
	v := validateRegexp(`^[a-z]+@[a-z]+\.[a-z]+$`)

	if err := v.Check("invalid"); err == nil {
		t.Error("validateRegexp should fail for 'invalid'")
	}
	if err := v.Check("test@example.com"); err != nil {
		t.Errorf("validateRegexp should pass for 'test@example.com': %v", err)
	}

	// Non-string should pass
	if err := v.Check(123); err != nil {
		t.Errorf("validateRegexp should pass for non-string: %v", err)
	}
}

func TestValidateOneOf(t *testing.T) {
	v := validateOneOf([]string{"debug", "info", "warn", "error"})

	if err := v.Check("invalid"); err == nil {
		t.Error("validateOneOf should fail for 'invalid'")
	}
	if err := v.Check("debug"); err != nil {
		t.Errorf("validateOneOf should pass for 'debug': %v", err)
	}
	if err := v.Check("info"); err != nil {
		t.Errorf("validateOneOf should pass for 'info': %v", err)
	}
	if err := v.Check("error"); err != nil {
		t.Errorf("validateOneOf should pass for 'error': %v", err)
	}

	// Non-string should pass
	if err := v.Check(123); err != nil {
		t.Errorf("validateOneOf should pass for non-string: %v", err)
	}
}

func TestValidateMinDuration(t *testing.T) {
	v := validateMinDuration(5 * time.Second)

	if err := v.Check(2 * time.Second); err == nil {
		t.Error("validateMinDuration(5s) should fail for 2s")
	}
	if err := v.Check(5 * time.Second); err != nil {
		t.Errorf("validateMinDuration(5s) should pass for 5s: %v", err)
	}
	if err := v.Check(10 * time.Second); err != nil {
		t.Errorf("validateMinDuration(5s) should pass for 10s: %v", err)
	}

	// Non-duration should pass
	if err := v.Check("string"); err != nil {
		t.Errorf("validateMinDuration should pass for non-duration: %v", err)
	}
}

func TestValidateMaxDuration(t *testing.T) {
	v := validateMaxDuration(1 * time.Minute)

	if err := v.Check(2 * time.Minute); err == nil {
		t.Error("validateMaxDuration(1m) should fail for 2m")
	}
	if err := v.Check(1 * time.Minute); err != nil {
		t.Errorf("validateMaxDuration(1m) should pass for 1m: %v", err)
	}
	if err := v.Check(30 * time.Second); err != nil {
		t.Errorf("validateMaxDuration(1m) should pass for 30s: %v", err)
	}
}

func TestValidateMinItems(t *testing.T) {
	v := validateMinItems(2)

	// String slice tests
	if err := v.Check([]string{"a"}); err == nil {
		t.Error("validateMinItems(2) should fail for []string with 1 item")
	}
	if err := v.Check([]string{"a", "b"}); err != nil {
		t.Errorf("validateMinItems(2) should pass for []string with 2 items: %v", err)
	}
	if err := v.Check([]string{"a", "b", "c"}); err != nil {
		t.Errorf("validateMinItems(2) should pass for []string with 3 items: %v", err)
	}

	// Int64 slice tests
	if err := v.Check([]int64{1}); err == nil {
		t.Error("validateMinItems(2) should fail for []int64 with 1 item")
	}
	if err := v.Check([]int64{1, 2}); err != nil {
		t.Errorf("validateMinItems(2) should pass for []int64 with 2 items: %v", err)
	}

	// Non-slice should pass
	if err := v.Check("string"); err != nil {
		t.Errorf("validateMinItems should pass for non-slice: %v", err)
	}
}

func TestValidateMaxItems(t *testing.T) {
	v := validateMaxItems(3)

	// String slice tests
	if err := v.Check([]string{"a", "b", "c", "d"}); err == nil {
		t.Error("validateMaxItems(3) should fail for []string with 4 items")
	}
	if err := v.Check([]string{"a", "b", "c"}); err != nil {
		t.Errorf("validateMaxItems(3) should pass for []string with 3 items: %v", err)
	}
	if err := v.Check([]string{"a", "b"}); err != nil {
		t.Errorf("validateMaxItems(3) should pass for []string with 2 items: %v", err)
	}

	// Int64 slice tests
	if err := v.Check([]int64{1, 2, 3, 4}); err == nil {
		t.Error("validateMaxItems(3) should fail for []int64 with 4 items")
	}
	if err := v.Check([]int64{1, 2, 3}); err != nil {
		t.Errorf("validateMaxItems(3) should pass for []int64 with 3 items: %v", err)
	}
}

func TestValidationName(t *testing.T) {
	// Test that validation names are set correctly
	tests := []struct {
		validation Validation
		contains   string
	}{
		{validateRequired(), "required"},
		{validateMin(10), "min(10)"},
		{validateMax(100), "max(100)"},
		{validateMinLength(5), "minLength(5)"},
		{validateMaxLength(10), "maxLength(10)"},
		{validateRegexp(`\d+`), "regexp"},
		{validateOneOf([]string{"a", "b"}), "oneOf"},
		{validateMinDuration(5 * time.Second), "minDuration"},
		{validateMaxDuration(1 * time.Minute), "maxDuration"},
		{validateMinItems(2), "minItems(2)"},
		{validateMaxItems(5), "maxItems(5)"},
	}

	for _, tt := range tests {
		if tt.validation.Name == "" {
			t.Errorf("Validation name should not be empty")
		}
	}
}
