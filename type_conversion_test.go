package commandkit

import (
	"testing"
	"time"
)

// TestTypeConversion_Int64 tests various conversions to int64
func TestTypeConversion_Int64(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  int64
		shouldErr bool
	}{
		{"int to int64", 5, 5, false},
		{"int32 to int64", int32(100), 100, false},
		{"int16 to int64", int16(50), 50, false},
		{"int8 to int64", int8(25), 25, false},
		{"uint to int64", uint(123), 123, false},
		{"uint32 to int64", uint32(456), 456, false},
		{"uint16 to int64", uint16(789), 789, false},
		{"uint8 to int64", uint8(200), 200, false},
		{"float64 to int64", 42.7, 42, false},
		{"float32 to int64", float32(33.3), 33, false},
		{"string to int64", "12345", 12345, false},
		{"bool true to int64", true, 1, false},
		{"bool false to int64", false, 0, false},
		{"invalid string to int64", "abc", 0, true},
		{"overflow uint to int64", uint(1 << 63), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.Define("TEST").Int64().Default(tt.value)

			err := cfg.Execute([]string{"test"})
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for conversion %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")
			value, err := Get[int64](ctx, "TEST")
			if err != nil {
				t.Errorf("Failed to get value: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

// TestTypeConversion_Float64 tests various conversions to float64
func TestTypeConversion_Float64(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  float64
		shouldErr bool
	}{
		{"int to float64", 5, 5.0, false},
		{"int64 to float64", int64(123), 123.0, false},
		{"int32 to float64", int32(45), 45.0, false},
		{"uint to float64", uint(67), 67.0, false},
		{"uint64 to float64", uint64(89), 89.0, false},
		{"float32 to float64", float32(3.14), 3.140000104904175, false}, // precision expected
		{"string to float64", "123.45", 123.45, false},
		{"bool true to float64", true, 1.0, false},
		{"bool false to float64", false, 0.0, false},
		{"invalid string to float64", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.Define("TEST").Float64().Default(tt.value)

			err := cfg.Execute([]string{"test"})
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for conversion %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")
			value, err := Get[float64](ctx, "TEST")
			if err != nil {
				t.Errorf("Failed to get value: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, value)
			}
		})
	}
}

// TestTypeConversion_String tests various conversions to string
func TestTypeConversion_String(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"int to string", 123, "123"},
		{"int64 to string", int64(456), "456"},
		{"float64 to string", 78.9, "78.9"},
		{"bool true to string", true, "true"},
		{"bool false to string", false, "false"},
		{"string to string", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.Define("TEST").String().Default(tt.value)

			err := cfg.Execute([]string{"test"})
			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")
			value, err := Get[string](ctx, "TEST")
			if err != nil {
				t.Errorf("Failed to get value: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, value)
			}
		})
	}
}

// TestTypeConversion_Bool tests various conversions to bool
func TestTypeConversion_Bool(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  bool
		shouldErr bool
	}{
		{"string true to bool", "true", true, false},
		{"string false to bool", "false", false, false},
		{"string 1 to bool", "1", true, false},
		{"string 0 to bool", "0", false, false},
		{"string yes to bool", "yes", true, false},
		{"string no to bool", "no", false, false},
		{"string on to bool", "on", true, false},
		{"string off to bool", "off", false, false},
		{"int non-zero to bool", 5, true, false},
		{"int zero to bool", 0, false, false},
		{"int64 non-zero to bool", int64(10), true, false},
		{"int64 zero to bool", int64(0), false, false},
		{"float64 non-zero to bool", 3.14, true, false},
		{"float64 zero to bool", 0.0, false, false},
		{"invalid string to bool", "maybe", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.Define("TEST").Bool().Default(tt.value)

			err := cfg.Execute([]string{"test"})
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for conversion %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")
			value, err := Get[bool](ctx, "TEST")
			if err != nil {
				t.Errorf("Failed to get value: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, value)
			}
		})
	}
}

// TestTypeConversion_Duration tests various conversions to time.Duration
func TestTypeConversion_Duration(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  time.Duration
		shouldErr bool
	}{
		{"string to duration", "5s", 5 * time.Second, false},
		{"string complex duration", "1h30m", time.Hour + 30*time.Minute, false},
		{"int to duration (seconds)", 120, 120 * time.Second, false},
		{"int64 to duration (seconds)", int64(300), 300 * time.Second, false},
		{"float64 to duration (seconds)", 45.5, 45 * time.Second, false},
		{"float32 to duration (seconds)", float32(10.5), 10 * time.Second, false},
		{"invalid string to duration", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.Define("TEST").Duration().Default(tt.value)

			err := cfg.Execute([]string{"test"})
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for conversion %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")
			value, err := Get[time.Duration](ctx, "TEST")
			if err != nil {
				t.Errorf("Failed to get value: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, value)
			}
		})
	}
}

// TestTypeConversion_Uint8 tests various conversions to uint8 with range validation
func TestTypeConversion_Uint8(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  uint8
		shouldErr bool
	}{
		{"int in range to uint8", 123, 123, false},
		{"int64 in range to uint8", int64(200), 200, false},
		{"string in range to uint8", "250", 250, false},
		{"int out of range to uint8", 300, 0, true},
		{"int64 out of range to uint8", int64(300), 0, true},
		{"string out of range to uint8", "300", 0, true},
		{"negative int to uint8", -5, 0, true},
		{"invalid string to uint8", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.Define("TEST").Uint8().Default(tt.value)

			err := cfg.Execute([]string{"test"})
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for conversion %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")
			value, err := Get[uint8](ctx, "TEST")
			if err != nil {
				t.Errorf("Failed to get value: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

// TestTypeConversion_Float32 tests various conversions to float32
func TestTypeConversion_Float32(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  float32
		shouldErr bool
	}{
		{"int to float32", 42, 42.0, false},
		{"int64 to float32", int64(123), 123.0, false},
		{"float64 to float32", 3.14159, 3.14159, false}, // precision expected
		{"string to float32", "2.718", 2.718, false},
		{"bool true to float32", true, 1.0, false},
		{"bool false to float32", false, 0.0, false},
		{"invalid string to float32", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.Define("TEST").Float32().Default(tt.value)

			err := cfg.Execute([]string{"test"})
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for conversion %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")
			value, err := Get[float32](ctx, "TEST")
			if err != nil {
				t.Errorf("Failed to get value: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, value)
			}
		})
	}
}

// TestTypeConversion_Slices tests various conversions to slice types
func TestTypeConversion_Slices(t *testing.T) {
	tests := []struct {
		name       string
		targetType string
		value      any
		expected   interface{}
		shouldErr  bool
	}{
		{
			"int slice to string slice",
			"StringSlice",
			[]int{1, 2, 3},
			[]string{"1", "2", "3"},
			false,
		},
		{
			"int64 slice to string slice",
			"StringSlice",
			[]int64{10, 20},
			[]string{"10", "20"},
			false,
		},
		{
			"bool slice to string slice",
			"StringSlice",
			[]bool{true, false},
			[]string{"true", "false"},
			false,
		},
		{
			"string slice to int64 slice",
			"Int64Slice",
			[]string{"100", "200"},
			[]int64{100, 200},
			false,
		},
		{
			"int slice to int64 slice",
			"Int64Slice",
			[]int{5, 10},
			[]int64{5, 10},
			false,
		},
		{
			"invalid string slice to int64 slice",
			"Int64Slice",
			[]string{"abc", "def"},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()

			// Define based on target type
			switch tt.targetType {
			case "StringSlice":
				cfg.Define("TEST").StringSlice().Default(tt.value)
			case "Int64Slice":
				cfg.Define("TEST").Int64Slice().Default(tt.value)
			}

			err := cfg.Execute([]string{"test"})
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for conversion %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected config processing error: %v", err)
			}

			ctx := NewCommandContext([]string{}, cfg, "test", "")

			// Use type assertion to handle different slice types
			switch tt.expected.(type) {
			case []string:
				value, err := Get[[]string](ctx, "TEST")
				if err != nil {
					t.Errorf("Failed to get value: %v", err)
				}
				if !equalStringSlices(value, tt.expected.([]string)) {
					t.Errorf("Expected %v, got %v", tt.expected, value)
				}
			case []int64:
				value, err := Get[[]int64](ctx, "TEST")
				if err != nil {
					t.Errorf("Failed to get value: %v", err)
				}
				if !equalInt64Slices(value, tt.expected.([]int64)) {
					t.Errorf("Expected %v, got %v", tt.expected, value)
				}
			}
		})
	}
}

// Helper functions for slice comparison
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalInt64Slices(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestTypeConversion_RealWorldScenario tests the original bug scenario
func TestTypeConversion_RealWorldScenario(t *testing.T) {
	// This is the exact scenario from the bug report
	cfg := New()

	// This should now work without explicit casting
	cfg.Define("SFTP_MAX_SESSIONS_PER_USER").Int64().Env("SFTP_MAX_SESSIONS_PER_USER").Min(1).Max(1000).Default(5)

	err := cfg.Execute([]string{"test"})
	if err != nil {
		t.Fatalf("Config processing failed: %v", err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")
	value, err := Get[int64](ctx, "SFTP_MAX_SESSIONS_PER_USER")
	if err != nil {
		t.Errorf("Failed to get SFTP_MAX_SESSIONS_PER_USER: %v", err)
	}

	if value != 5 {
		t.Errorf("Expected 5, got %d", value)
	}

	// Test that validation still works
	cfg2 := New()
	cfg2.Define("TEST").Uint8().Default(300) // 300 > uint8 max (255)

	err = cfg2.Execute([]string{"test"})
	if err == nil {
		t.Error("Expected error for out-of-range conversion")
	}
}
