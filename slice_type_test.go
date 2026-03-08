// commandkit/slice_type_test.go
package commandkit

import (
	"testing"
)

func TestSliceTypeNoPanic(t *testing.T) {
	// Test that slice types don't cause panic in typeDescription
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("typeDescription panicked with slice types: %v", r)
		}
	}()

	// Test various slice types
	sliceTests := []struct {
		name     string
		value    any
		expected string
	}{
		{"string slice", []string{"a", "b"}, "[]string"},
		{"int64 slice", []int64{1, 2}, "[]int64"},
		{"int slice", []int{1, 2}, "[]int"},
		{"empty string slice", []string{}, "[]string"},
		{"empty int64 slice", []int64{}, "[]int64"},
		{"empty int slice", []int{}, "[]int"},
	}

	for _, tt := range sliceTests {
		t.Run(tt.name, func(t *testing.T) {
			result := typeDescription(tt.value)
			if result != tt.expected {
				t.Errorf("typeDescription(%v) = %q, expected %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestInt64Consistency(t *testing.T) {
	// Test that int64 always returns "int64" never "int"
	int64Value := int64(123)
	result := typeDescription(int64Value)
	if result != "int64" {
		t.Errorf("typeDescription(int64) = %q, expected \"int64\"", result)
	}

	// Test that int returns "int"
	intValue := int(123)
	result = typeDescription(intValue)
	if result != "int" {
		t.Errorf("typeDescription(int) = %q, expected \"int\"", result)
	}
}

func TestTypeCachingWorks(t *testing.T) {
	// Test that type caching still works for basic types
	basicTests := []struct {
		name     string
		value    any
		expected string
	}{
		{"string", "test", "string"},
		{"int64", int64(123), "int64"},
		{"int", int(123), "int"},
		{"bool", true, "bool"},
		{"float64", 3.14, "float64"},
	}

	for _, tt := range basicTests {
		t.Run(tt.name, func(t *testing.T) {
			// Call multiple times to test caching
			result1 := typeDescription(tt.value)
			result2 := typeDescription(tt.value)

			if result1 != tt.expected {
				t.Errorf("First call typeDescription(%v) = %q, expected %q", tt.value, result1, tt.expected)
			}
			if result2 != tt.expected {
				t.Errorf("Second call typeDescription(%v) = %q, expected %q", tt.value, result2, tt.expected)
			}
			if result1 != result2 {
				t.Errorf("Cached results differ: first=%q, second=%q", result1, result2)
			}
		})
	}
}

func TestSliceTypeCachingWorks(t *testing.T) {
	// Test that type caching works for slice types
	sliceValue := []string{"a", "b"}

	// Call multiple times to test caching
	result1 := typeDescription(sliceValue)
	result2 := typeDescription(sliceValue)

	expected := "[]string"
	if result1 != expected {
		t.Errorf("First call typeDescription(%v) = %q, expected %q", sliceValue, result1, expected)
	}
	if result2 != expected {
		t.Errorf("Second call typeDescription(%v) = %q, expected %q", sliceValue, result2, expected)
	}
	if result1 != result2 {
		t.Errorf("Cached slice results differ: first=%q, second=%q", result1, result2)
	}
}

func TestIntTypeSupport(t *testing.T) {
	cfg := New()

	// Test Int() method
	cfg.Define("INT_PORT").
		Int().
		Flag("port").
		Default(8080)

	// Test IntSlice() method
	cfg.Define("INT_PORTS").
		IntSlice().
		Flag("ports").
		Default([]int{8080, 8081})

	// Verify the types are set correctly
	if def, exists := cfg.definitions["INT_PORT"]; exists {
		if def.valueType != TypeInt {
			t.Errorf("INT_PORT type = %v, expected TypeInt", def.valueType)
		}
	} else {
		t.Error("INT_PORT definition not found")
	}

	if def, exists := cfg.definitions["INT_PORTS"]; exists {
		if def.valueType != TypeIntSlice {
			t.Errorf("INT_PORTS type = %v, expected TypeIntSlice", def.valueType)
		}
	} else {
		t.Error("INT_PORTS definition not found")
	}
}

// TestSliceTypeRetrievalFixed tests that the slice regression is fixed
func TestSliceTypeRetrievalFixed(t *testing.T) {
	cfg := New()

	cfg.Define("TAGS").
		StringSlice().
		Default([]string{"ssh-rsa", "ssh-ed25519"})

	cfg.Define("NUMBERS").
		Int64Slice().
		Default([]int64{1, 2, 3})

	cfg.Define("PORTS").
		IntSlice().
		Default([]int{8080, 8081})

	// Test configuration processing
	err := cfg.Execute([]string{"test"})
	if err != nil {
		t.Fatalf("Config execution failed: %v", err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test Get[[]string] - should work without workaround
	tags, err := Get[[]string](ctx, "TAGS")
	if err != nil {
		t.Fatalf("Get[[]string] failed: %v", err)
	}
	expectedTags := []string{"ssh-rsa", "ssh-ed25519"}
	if len(tags) != len(expectedTags) {
		t.Fatalf("Expected %d tags, got %d", len(expectedTags), len(tags))
	}
	for i, tag := range expectedTags {
		if tags[i] != tag {
			t.Errorf("Expected tag[%d] = %q, got %q", i, tag, tags[i])
		}
	}

	// Test Get[[]int64]
	numbers, err := Get[[]int64](ctx, "NUMBERS")
	if err != nil {
		t.Fatalf("Get[[]int64] failed: %v", err)
	}
	expectedNumbers := []int64{1, 2, 3}
	if len(numbers) != len(expectedNumbers) {
		t.Fatalf("Expected %d numbers, got %d", len(expectedNumbers), len(numbers))
	}
	for i, num := range expectedNumbers {
		if numbers[i] != num {
			t.Errorf("Expected number[%d] = %d, got %d", i, num, numbers[i])
		}
	}

	// Test Get[[]int]
	ports, err := Get[[]int](ctx, "PORTS")
	if err != nil {
		t.Fatalf("Get[[]int] failed: %v", err)
	}
	expectedPorts := []int{8080, 8081}
	if len(ports) != len(expectedPorts) {
		t.Fatalf("Expected %d ports, got %d", len(expectedPorts), len(ports))
	}
	for i, port := range expectedPorts {
		if ports[i] != port {
			t.Errorf("Expected port[%d] = %d, got %d", i, port, ports[i])
		}
	}
}

// TestNoSliceWorkaroundNeeded ensures users don't need workarounds
func TestNoSliceWorkaroundNeeded(t *testing.T) {
	cfg := New()

	cfg.Define("ALGORITHMS").
		StringSlice().
		Default([]string{"ssh-rsa", "ssh-ed25519"})

	err := cfg.Execute([]string{"test"})
	if err != nil {
		t.Fatalf("Config execution failed: %v", err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Should work directly without string splitting workaround
	algorithms, err := Get[[]string](ctx, "ALGORITHMS")
	if err != nil {
		t.Fatalf("Get[[]string] failed: %v", err)
	}

	expected := []string{"ssh-rsa", "ssh-ed25519"}
	if len(algorithms) != len(expected) {
		t.Fatalf("Expected %d algorithms, got %d", len(expected), len(algorithms))
	}
	for i, algo := range expected {
		if algorithms[i] != algo {
			t.Errorf("Expected algorithm[%d] = %q, got %q", i, algo, algorithms[i])
		}
	}

	// Should NOT return string representation when requesting string
	_, err = Get[string](ctx, "ALGORITHMS")
	if err == nil {
		t.Error("Expected type mismatch error when getting string instead of []string")
	}
}

// TestAllSourcesSliceConsistency tests slice consistency across all value sources
func TestAllSourcesSliceConsistency(t *testing.T) {
	// Test each source type consistently
	testCases := []struct {
		name     string
		args     []string
		expected []string
		env      string
	}{
		{"default", []string{}, []string{"default1", "default2"}, ""},
		{"env", []string{}, []string{"env1", "env2", "env3"}, "env1,env2,env3"},
		{"flag", []string{"--tags", "flag1,flag2"}, []string{"flag1", "flag2"}, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for this test
			if tc.env != "" {
				t.Setenv("TEST_TAGS", tc.env)
			}

			// Create fresh config for each test
			testCfg := New()
			testCfg.Define("TAGS").
				StringSlice().
				Env("TEST_TAGS").
				Flag("tags").
				Default([]string{"default1", "default2"})

			err := testCfg.Execute(append([]string{"test"}, tc.args...))
			if err != nil {
				t.Fatalf("Config execution failed: %v", err)
			}

			ctx := NewCommandContext([]string{}, testCfg, "test", "")
			tags, err := Get[[]string](ctx, "TAGS")
			if err != nil {
				t.Fatalf("Get[[]string] failed: %v", err)
			}

			if len(tags) != len(tc.expected) {
				t.Fatalf("Expected %d tags, got %d", len(tc.expected), len(tags))
			}
			for i, tag := range tc.expected {
				if tags[i] != tag {
					t.Errorf("Expected tag[%d] = %q, got %q", i, tag, tags[i])
				}
			}
		})
	}
}

func TestValueTypeStringMethods(t *testing.T) {
	tests := []struct {
		valueType ValueType
		expected  string
	}{
		{TypeString, "string"},
		{TypeInt64, "int64"},
		{TypeInt, "int"},
		{TypeFloat64, "float64"},
		{TypeBool, "bool"},
		{TypeDuration, "duration"},
		{TypeURL, "url"},
		{TypeStringSlice, "[]string"},
		{TypeInt64Slice, "[]int64"},
		{TypeIntSlice, "[]int"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.valueType.String()
			if result != tt.expected {
				t.Errorf("ValueType(%v).String() = %q, expected %q", tt.valueType, result, tt.expected)
			}
		})
	}
}
