package commandkit

import (
	"testing"
)

// BenchmarkCachedTemplateRenderer benchmarks the new cached template renderer
func BenchmarkCachedTemplateRenderer(b *testing.B) {
	renderer := NewCachedTemplateRenderer()

	templateStr := `Usage: {{.Executable}} [options]

{{if .Commands}}
Commands:
{{range .Commands}}
  {{.Name}} - {{.ShortHelp}}
{{end}}
{{end}}

{{if .Flags}}
Flags:
{{range .Flags}}
  --{{.Name}} {{.Type}} (default: {{.Default}})
{{end}}
{{end}}`

	data := map[string]interface{}{
		"Executable": "test",
		"Commands": []map[string]interface{}{
			{"Name": "start", "ShortHelp": "Start the service"},
			{"Name": "stop", "ShortHelp": "Stop the service"},
			{"Name": "status", "ShortHelp": "Show status"},
			{"Name": "restart", "ShortHelp": "Restart the service"},
			{"Name": "deploy", "ShortHelp": "Deploy the application"},
		},
		"Flags": []map[string]interface{}{
			{"Name": "port", "Type": "int64", "Default": "8080"},
			{"Name": "host", "Type": "string", "Default": "localhost"},
			{"Name": "debug", "Type": "bool", "Default": "false"},
			{"Name": "timeout", "Type": "duration", "Default": "30s"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.Render(templateStr, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOriginalTemplateRenderer benchmarks the original template renderer
func BenchmarkOriginalTemplateRenderer(b *testing.B) {
	renderer := NewGoTemplateRenderer()

	templateStr := `Usage: {{.Executable}} [options]

{{if .Commands}}
Commands:
{{range .Commands}}
  {{.Name}} - {{.ShortHelp}}
{{end}}
{{end}}

{{if .Flags}}
Flags:
{{range .Flags}}
  --{{.Name}} {{.Type}} (default: {{.Default}})
{{end}}
{{end}}`

	data := map[string]interface{}{
		"Executable": "test",
		"Commands": []map[string]interface{}{
			{"Name": "start", "ShortHelp": "Start the service"},
			{"Name": "stop", "ShortHelp": "Stop the service"},
			{"Name": "status", "ShortHelp": "Show status"},
			{"Name": "restart", "ShortHelp": "Restart the service"},
			{"Name": "deploy", "ShortHelp": "Deploy the application"},
		},
		"Flags": []map[string]interface{}{
			{"Name": "port", "Type": "int64", "Default": "8080"},
			{"Name": "host", "Type": "string", "Default": "localhost"},
			{"Name": "debug", "Type": "bool", "Default": "false"},
			{"Name": "timeout", "Type": "duration", "Default": "30s"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.Render(templateStr, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHelpFormatter_Cached benchmarks help formatter with cached renderer
func BenchmarkHelpFormatter_Cached(b *testing.B) {
	formatter := NewTemplateHelpFormatter()

	// Create a complex help structure
	help := &GlobalHelp{
		Executable: "testapp",
		Commands: []CommandSummary{
			{Name: "start", Description: "Start the service"},
			{Name: "stop", Description: "Stop the service"},
			{Name: "status", Description: "Show status"},
			{Name: "restart", Description: "Restart the service"},
			{Name: "deploy", Description: "Deploy the application"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := formatter.FormatGlobalHelp(help)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTemplateRenderer_Concurrent benchmarks concurrent template rendering
func BenchmarkTemplateRenderer_Concurrent(b *testing.B) {
	renderer := NewCachedTemplateRenderer()

	templateStr := `Usage: {{.Executable}} [options]

Commands:
{{range .Commands}}
  {{.Name}} - {{.ShortHelp}}
{{end}}`

	data := map[string]interface{}{
		"Executable": "test",
		"Commands": []map[string]interface{}{
			{"Name": "start", "ShortHelp": "Start the service"},
			{"Name": "stop", "ShortHelp": "Stop the service"},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := renderer.Render(templateStr, data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
