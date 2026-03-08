package commandkit

import (
	"testing"
	"time"
)

// BenchmarkConfigProcessing_Large benchmarks configuration processing with many definitions
func BenchmarkConfigProcessing_Large(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := New()

		// Create many definitions to simulate real-world usage
		for j := 0; j < 100; j++ {
			cfg.Define("PORT").
				Int64().
				Env("PORT").
				Flag("port").
				Default(int64(8080))

			cfg.Define("HOST").
				String().
				Env("HOST").
				Flag("host").
				Default("localhost")

			cfg.Define("DEBUG").
				Bool().
				Env("DEBUG").
				Flag("debug").
				Default(false)

			cfg.Define("RATE").
				Float64().
				Env("RATE").
				Flag("rate").
				Default(100.0).
				Range(1.0, 1000.0)

			cfg.Define("TIMEOUT").
				Duration().
				Env("TIMEOUT").
				Flag("timeout").
				Default(30 * time.Second)
		}

		// Process configuration
		if err := cfg.Execute([]string{"test"}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConfigProcessing_Small benchmarks configuration processing with few definitions
func BenchmarkConfigProcessing_Small(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := New()

		cfg.Define("PORT").
			Int64().
			Env("PORT").
			Flag("port").
			Default(int64(8080))

		cfg.Define("HOST").
			String().
			Env("HOST").
			Flag("host").
			Default("localhost")

		// Process configuration
		if err := cfg.Execute([]string{"test"}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHelpGeneration_Global benchmarks global help generation performance
func BenchmarkHelpGeneration_Global(b *testing.B) {
	cfg := New()

	// Create multiple commands
	for i := 0; i < 10; i++ {
		cfg.Command("start").
			Func(func(ctx *CommandContext) error { return nil }).
			ShortHelp("Start the service").
			LongHelp("Start the service with all components initialized.")

		cfg.Command("stop").
			Func(func(ctx *CommandContext) error { return nil }).
			ShortHelp("Stop the service").
			LongHelp("Stop the service gracefully.")

		cfg.Command("status").
			Func(func(ctx *CommandContext) error { return nil }).
			ShortHelp("Show service status").
			LongHelp("Display current service status and statistics.")
	}

	// Process configuration
	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate help
		help := cfg.GenerateHelp()
		if help == "" {
			b.Fatal("Empty help generated")
		}
	}
}

// BenchmarkHelpGeneration_Command benchmarks command-specific help generation
func BenchmarkHelpGeneration_Command(b *testing.B) {
	cfg := New()

	// Create a command with many definitions
	cfg.Command("deploy").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Deploy the application").
		LongHelp("Deploy the application to the specified environment.").
		Config(func(cc *CommandConfig) {
			for i := 0; i < 20; i++ {
				cc.Define("PORT").
					Int64().
					Flag("port").
					Default(int64(8080))

				cc.Define("ENV").
					String().
					Flag("env").
					Default("development")

				cc.Define("DRY_RUN").
					Bool().
					Flag("dry-run").
					Default(false)
			}
		})

	// Process configuration
	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate command help
		err := cfg.ShowCommandHelp("deploy")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTemplateRendering_Cached benchmarks template rendering with caching
func BenchmarkTemplateRendering_Cached(b *testing.B) {
	renderer := NewGoTemplateRenderer()

	templateStr := `Usage: {{.Executable}} [options]

Options:
{{range .Commands}}
  {{.Name}} - {{.ShortHelp}}
{{end}}`

	data := map[string]interface{}{
		"Executable": "test",
		"Commands": []map[string]interface{}{
			{"Name": "start", "ShortHelp": "Start the service"},
			{"Name": "stop", "ShortHelp": "Stop the service"},
			{"Name": "status", "ShortHelp": "Show status"},
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

// BenchmarkTemplateRendering_Uncached benchmarks template rendering without caching
func BenchmarkTemplateRendering_Uncached(b *testing.B) {
	templateStr := `Usage: {{.Executable}} [options]

Options:
{{range .Commands}}
  {{.Name}} - {{.ShortHelp}}
{{end}}`

	data := map[string]interface{}{
		"Executable": "test",
		"Commands": []map[string]interface{}{
			{"Name": "start", "ShortHelp": "Start the service"},
			{"Name": "stop", "ShortHelp": "Stop the service"},
			{"Name": "status", "ShortHelp": "Show status"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer := NewGoTemplateRenderer()
		_, err := renderer.Render(templateStr, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidation_MultipleTypes benchmarks validation with different types
func BenchmarkValidation_MultipleTypes(b *testing.B) {
	cfg := New()

	// Define various validation types
	cfg.Define("PORT").
		Int64().
		Range(1, 65535).
		Default(8080)

	cfg.Define("EMAIL").
		String().
		Regexp(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
		Default("test@example.com")

	cfg.Define("RATE").
		Float64().
		Range(0.1, 100.0).
		Default(1.0)

	cfg.Define("TOKEN").
		String().
		MinLength(32).
		MaxLength(64).
		Default("12345678901234567890123456789012")

	cfg.Define("ENVIRONMENTS").
		StringSlice().
		ItemsRange(1, 5).
		Default([]string{"dev", "staging", "prod"})

	cfg.Define("TIMEOUT").
		Duration().
		DurationRange(1*time.Second, 5*time.Minute).
		Default(30 * time.Second)

	// Process configuration
	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get all values to trigger validation
		_, _ = Get[int64](ctx, "PORT")
		_, _ = Get[string](ctx, "EMAIL")
		_, _ = Get[float64](ctx, "RATE")
		_, _ = Get[string](ctx, "TOKEN")
		_, _ = Get[[]string](ctx, "ENVIRONMENTS")
		_, _ = Get[time.Duration](ctx, "TIMEOUT")
	}
}

// BenchmarkErrorFormatting benchmarks error message formatting
func BenchmarkErrorFormatting(b *testing.B) {
	cfg := New()

	cfg.Define("PORT").
		Int64().
		Range(1, 65535).
		Default(8080)

	cfg.Define("EMAIL").
		String().
		Regexp(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
		Default("test@example.com")

	// Process configuration
	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Create execution context with errors
	execCtx := NewExecutionContext("test")

	// Add some config errors
	configErr := ConfigError{
		Key:              "PORT",
		Source:           "flag",
		Value:            "99999",
		ErrorDescription: "value 99999 is greater than maximum 65535",
	}

	execCtx.CollectConfigError(cfg, configErr)

	ctx.execution = execCtx

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Format error with template
		_, err := ctx.execution.renderErrorsWithCommand(nil, cfg.getHelpService())
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFileOperations benchmarks file loading and processing
func BenchmarkFileOperations(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := New()
		cfg.SetDefaultPriority(PriorityFileEnvFlagDefault)

		cfg.Define("PORT").
			Int64().
			File("port_in_file").
			Default(8080)

		cfg.Define("HOST").
			String().
			File("host_in_file").
			Default("localhost")

		cfg.Define("DEBUG").
			Bool().
			File("debug_in_file").
			Default(false)

		cfg.Define("RATE").
			Float64().
			File("rate_in_file").
			Default(100.0)

		cfg.Define("TIMEOUT").
			Duration().
			File("timeout_in_file").
			Default(30 * time.Second)

		// Simulate file loading (in real usage this would be cfg.LoadFile())
		if cfg.fileConfig == nil {
			cfg.fileConfig = &FileConfig{
				data: map[string]any{
					"port_in_file":  3000.0,
					"host_in_file":  "localhost",
					"debug_in_file": true,
					"rate_in_file":  100.5,
					"environments": map[string]any{
						"development": map[string]any{
							"timeout_in_file": "30s",
						},
						"production": map[string]any{
							"timeout_in_file": "10s",
						},
					},
				},
			}
		}

		// Process configuration
		if err := cfg.Execute([]string{"test"}); err != nil {
			b.Fatal(err)
		}
	}
}
