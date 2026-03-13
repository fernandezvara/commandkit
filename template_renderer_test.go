// commandkit/template_renderer_test.go
package commandkit

import (
	"strings"
	"testing"
	"text/template"
)

func TestNewGoTemplateRenderer(t *testing.T) {
	renderer := NewGoTemplateRenderer()
	if renderer == nil {
		t.Error("Expected non-nil renderer")
	}

	// Check default functions exist
	funcMap := renderer.GetFuncMap()
	if funcMap == nil {
		t.Error("Expected non-nil function map")
	}

	// Check some default functions
	if _, exists := funcMap["join"]; !exists {
		t.Error("Expected 'join' function in func map")
	}

	if _, exists := funcMap["upper"]; !exists {
		t.Error("Expected 'upper' function in func map")
	}
}

func TestGoTemplateRenderer_Render(t *testing.T) {
	renderer := NewGoTemplateRenderer()

	tests := []struct {
		name      string
		template  string
		data      any
		expected  string
		shouldErr bool
	}{
		{
			name:     "Simple template",
			template: "Hello, {{.Name}}!",
			data:     map[string]string{"Name": "World"},
			expected: "Hello, World!",
		},
		{
			name:     "Template with join function",
			template: "Items: {{join .Items \", \"}}",
			data:     map[string][]string{"Items": {"a", "b", "c"}},
			expected: "Items: a, b, c",
		},
		{
			name:     "Template with upper function",
			template: "Upper: {{.Text | upper}}",
			data:     map[string]string{"Text": "hello"},
			expected: "Upper: HELLO",
		},
		{
			name:     "Template with title function",
			template: "Title: {{.Text | title}}",
			data:     map[string]string{"Text": "hello world"},
			expected: "Title: Hello World",
		},
		{
			name:     "Template with format function",
			template: "Number: {{format \"value: %d\" .Number}}",
			data:     map[string]int{"Number": 42},
			expected: "Number: value: 42",
		},
		{
			name:      "Invalid template",
			template:  "Hello, {{.Name",
			data:      map[string]string{"Name": "World"},
			shouldErr: true,
		},
		{
			name:     "Missing variable",
			template: "Hello, {{.Missing}}!",
			data:     map[string]string{"Name": "World"},
			expected: "Hello, <no value>!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render(tt.template, tt.data)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGoTemplateRenderer_SetFuncMap(t *testing.T) {
	renderer := NewGoTemplateRenderer()

	// Create custom function map
	customFuncMap := template.FuncMap{
		"custom": func(s string) string {
			return "CUSTOM:" + s
		},
	}

	renderer.SetFuncMap(customFuncMap)

	// Test with custom function
	result, err := renderer.Render("Test: {{custom .Text}}", map[string]string{"Text": "hello"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Test: CUSTOM:hello"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGoTemplateRenderer_AddFunction(t *testing.T) {
	renderer := NewGoTemplateRenderer()

	// Add custom function
	renderer.AddFunction("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})

	// Test with added function
	result, err := renderer.Render("Reverse: {{reverse .Text}}", map[string]string{"Text": "hello"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Reverse: olleh"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGoTemplateRenderer_DefaultFunctions(t *testing.T) {
	renderer := NewGoTemplateRenderer()

	data := map[string]any{
		"Text":   "Hello World",
		"Items":  []string{"apple", "banana", "cherry"},
		"Number": 42,
		"Prefix": "test_",
		"Suffix": "_end",
	}

	template := `{{.Text | upper}}
{{.Text | lower}}
{{.Text | title}}
{{join .Items ", "}}
{{format "Number: %d" .Number}}
{{.Prefix | printf "%svalue"}}
{{contains .Text "World"}}
{{hasPrefix .Text "Hello"}}
{{hasSuffix .Text "World"}}`

	result, err := renderer.Render(template, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that all default functions work
	expectedParts := []string{
		"HELLO WORLD",
		"hello world",
		"Hello World",
		"apple, banana, cherry",
		"Number: 42",
		"test_value",
		"true",
		"true",
		"true",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain %q, got: %s", part, result)
		}
	}
}

func TestGoTemplateRenderer_ComplexTemplate(t *testing.T) {
	renderer := NewGoTemplateRenderer()

	data := map[string]any{
		"Commands": []map[string]any{
			{"Name": "start", "Description": "Start the service"},
			{"Name": "stop", "Description": "Stop the service"},
		},
		"Executable": "testapp",
	}

	template := `{{.Executable}} Commands:

{{range .Commands}}  {{.Name | printf "%-12s"}} {{.Description}}
{{end}}`

	result, err := renderer.Render(template, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "testapp Commands:\n\n  start        Start the service\n  stop         Stop the service\n"

	if result != expected {
		// Print actual result for debugging
		t.Logf("Actual result:\n%q\n", result)
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestGoTemplateRenderer_ErrorHandling(t *testing.T) {
	renderer := NewGoTemplateRenderer()

	tests := []struct {
		name     string
		template string
		data     any
	}{
		{
			name:     "Syntax error",
			template: "Hello, {{.Name",
			data:     map[string]string{"Name": "World"},
		},
		{
			name:     "Function error",
			template: "{{nonexistent .Field}}",
			data:     map[string]string{"Field": "value"},
		},
		{
			name:     "Type error",
			template: "{{.Number | upper}}",
			data:     map[string]int{"Number": 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := renderer.Render(tt.template, tt.data)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}
