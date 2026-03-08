package commandkit

import (
	"testing"
)

func TestTypeConverterConsistency(t *testing.T) {
	converter := NewTypeConverter()

	// Test all types produce consistent results
	testCases := []struct {
		name      string
		value     any
		expected  string
		delimiter string
		shouldErr bool
	}{
		{
			name:      "string",
			value:     "hello",
			expected:  "hello",
			delimiter: ",",
		},
		{
			name:      "bool true",
			value:     true,
			expected:  "true",
			delimiter: ",",
		},
		{
			name:      "bool false",
			value:     false,
			expected:  "false",
			delimiter: ",",
		},
		{
			name:      "int64",
			value:     int64(123),
			expected:  "123",
			delimiter: ",",
		},
		{
			name:      "int",
			value:     int(456),
			expected:  "456",
			delimiter: ",",
		},
		{
			name:      "float64",
			value:     3.14,
			expected:  "3.14",
			delimiter: ",",
		},
		{
			name:      "string slice",
			value:     []string{"a", "b", "c"},
			expected:  "a,b,c",
			delimiter: ",",
		},
		{
			name:      "string slice with pipe",
			value:     []string{"a", "b", "c"},
			expected:  "a|b|c",
			delimiter: "|",
		},
		{
			name:      "empty string slice",
			value:     []string{},
			expected:  "",
			delimiter: ",",
		},
		{
			name:      "int64 slice",
			value:     []int64{1, 2, 3},
			expected:  "1,2,3",
			delimiter: ",",
		},
		{
			name:      "int slice",
			value:     []int{4, 5, 6},
			expected:  "4,5,6",
			delimiter: ",",
		},
		{
			name:      "[]any slice",
			value:     []any{"x", 1, true},
			expected:  "x,1,true",
			delimiter: ",",
		},
		{
			name:      "unsupported type",
			value:     map[string]string{"key": "value"},
			expected:  "",
			delimiter: ",",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertToString(tc.value, tc.delimiter)

			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error for unsupported type %T", tc.value)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestTypeConverterDisplayString(t *testing.T) {
	converter := NewTypeConverter()

	testCases := []struct {
		name      string
		value     any
		expected  string
		delimiter string
	}{
		{
			name:      "supported type",
			value:     []string{"a", "b"},
			expected:  "a,b",
			delimiter: ",",
		},
		{
			name:      "unsupported type",
			value:     map[string]string{"key": "value"},
			expected:  "map[key:value]",
			delimiter: ",",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := converter.ConvertToDisplayString(tc.value, tc.delimiter)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestTypeConverterIsSupportedType(t *testing.T) {
	converter := NewTypeConverter()

	supportedTypes := []any{
		"string",
		true,
		false,
		int(123),
		int64(456),
		3.14,
		[]string{"a", "b"},
		[]int64{1, 2},
		[]int{3, 4},
		[]any{"x", 1},
	}

	for _, value := range supportedTypes {
		if !converter.IsSupportedType(value) {
			t.Errorf("Expected supported type: %T", value)
		}
	}

	unsupportedTypes := []any{
		map[string]string{"key": "value"},
		complex(1, 2),
		chan int(nil),
		func() {},
	}

	for _, value := range unsupportedTypes {
		if converter.IsSupportedType(value) {
			t.Errorf("Expected unsupported type: %T", value)
		}
	}
}
