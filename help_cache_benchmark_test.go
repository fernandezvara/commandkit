package commandkit

import (
	"testing"
)

// BenchmarkHelpGeneration_Cached benchmarks help generation with caching
func BenchmarkHelpGeneration_Cached(b *testing.B) {
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
		// Generate help (should be cached after first call)
		help := cfg.GenerateHelp()
		if help == "" {
			b.Fatal("Empty help generated")
		}
	}
}

// BenchmarkHelpGeneration_Uncached benchmarks help generation without caching
func BenchmarkHelpGeneration_Uncached(b *testing.B) {
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

	// Clear cache before benchmarking
	if cachedFormatter, ok := cfg.getHelpService().GetFormatter().(*CachedHelpFormatter); ok {
		cachedFormatter.ClearCache()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate help (uncached)
		help := cfg.GenerateHelp()
		if help == "" {
			b.Fatal("Empty help generated")
		}
	}
}

// BenchmarkCommandHelp_Cached benchmarks command help generation with caching
func BenchmarkCommandHelp_Cached(b *testing.B) {
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
					Default(8080)

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
		// Generate command help (should be cached after first call)
		err := cfg.ShowCommandHelp("deploy")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCommandHelp_Uncached benchmarks command help generation without caching
func BenchmarkCommandHelp_Uncached(b *testing.B) {
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
					Default(8080)

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

	// Clear cache before benchmarking
	if cachedFormatter, ok := cfg.getHelpService().GetFormatter().(*CachedHelpFormatter); ok {
		cachedFormatter.ClearCache()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate command help (uncached)
		err := cfg.ShowCommandHelp("deploy")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHelpFormatter_CachedComplex benchmarks help formatter with caching (complex)
func BenchmarkHelpFormatter_CachedComplex(b *testing.B) {
	formatter := NewTemplateHelpFormatter()
	cachedFormatter := NewCachedHelpFormatter(formatter)

	// Create a complex help structure
	help := &GlobalHelp{
		Executable: "testapp",
		Commands: []CommandSummary{
			{Name: "start", Description: "Start the service"},
			{Name: "stop", Description: "Stop the service"},
			{Name: "status", Description: "Show status"},
			{Name: "restart", Description: "Restart the service"},
			{Name: "deploy", Description: "Deploy the application"},
			{Name: "logs", Description: "Show logs"},
			{Name: "config", Description: "Manage configuration"},
			{Name: "health", Description: "Check health"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cachedFormatter.FormatGlobalHelp(help)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHelpFormatter_Uncached benchmarks help formatter without caching
func BenchmarkHelpFormatter_Uncached(b *testing.B) {
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
			{Name: "logs", Description: "Show logs"},
			{Name: "config", Description: "Manage configuration"},
			{Name: "health", Description: "Check health"},
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
