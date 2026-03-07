// examples/priority/priority_example.go
package main

import (
	"fmt"
	"os"

	"github.com/fernandezvara/commandkit"
)

func main() {
	fmt.Println("=== Configurable Source Priority Example ===\n")

	// Example 1: Default priority (Flag > Env > Default)
	fmt.Println("1. Default Priority (Flag > Env > Default):")
	cfg1 := commandkit.New()

	cfg1.Define("PORT").
		Int64().
		Env("PORT").
		Flag("port").
		Default(8080)

	// Set environment variable
	os.Setenv("PORT", "9000")

	// Simulate command line args
	originalArgs := os.Args
	os.Args = []string{"example", "--port", "3000"}

	result1 := cfg1.Process()
	if result1.Error != nil {
		fmt.Printf("Error: %v\n", result1.Error)
	} else {
		fmt.Printf("  PORT configured with Flag > Env > Default priority\n")
		fmt.Printf("  Flag value (3000) would win over Env (9000) and Default (8080)\n")
	}

	// Show override warnings
	if cfg1.HasOverrideWarnings() {
		warnings := cfg1.GetOverrideWarnings()
		for _, warning := range warnings.GetWarnings() {
			fmt.Printf("  Warning: %s -> %s (%s -> %s)\n",
				warning.Source, warning.OverrideBy, warning.OldValue, warning.NewValue)
		}
	}

	// Example 2: Custom priority (Env > Flag > Default)
	fmt.Println("\n2. Custom Priority (Env > Flag > Default):")
	cfg2 := commandkit.New()

	cfg2.Define("DATABASE_URL").
		String().
		Env("DATABASE_URL").
		Flag("database-url").
		Default("localhost:5432").
		PriorityEnvFlagDefault() // Env > Flag > Default

	// Set environment variable
	os.Setenv("DATABASE_URL", "postgres://prod:5432/mydb")

	result2 := cfg2.Process()
	if result2.Error != nil {
		fmt.Printf("Error: %v\n", result2.Error)
	} else {
		fmt.Printf("  DATABASE_URL configured with Env > Flag > Default priority\n")
		fmt.Printf("  Env value (postgres://prod:5432/mydb) would win over Flag and Default\n")
	}

	// Example 3: Config-level default priority
	fmt.Println("\n3. Config-level Default Priority (Env > Flag > Default):")
	cfg3 := commandkit.New()

	// Set config-level priority
	cfg3.SetDefaultPriority(commandkit.PriorityEnvFlagDefault)

	cfg3.Define("API_KEY").
		String().
		Env("API_KEY").
		Flag("api-key").
		Default("default-key")
		// No explicit priority - uses config default

	// Set environment variable
	os.Setenv("API_KEY", "prod-api-key")

	result3 := cfg3.Process()
	if result3.Error != nil {
		fmt.Printf("Error: %v\n", result3.Error)
	} else {
		fmt.Printf("  API_KEY configured with config-level Env > Flag > Default priority\n")
		fmt.Printf("  Env value (prod-api-key) would win due to config-level priority\n")
	}

	// Example 4: File-first priority
	fmt.Println("\n4. File-first Priority (File > Env > Flag > Default):")
	cfg4 := commandkit.New()

	cfg4.Define("LOG_LEVEL").
		String().
		Env("LOG_LEVEL").
		Flag("log-level").
		Default("info").
		PriorityFileEnvFlagDefault() // File > Env > Flag > Default

	result4 := cfg4.Process()
	if result4.Error != nil {
		fmt.Printf("Error: %v\n", result4.Error)
	} else {
		fmt.Printf("  LOG_LEVEL configured with File > Env > Flag > Default priority\n")
		fmt.Printf("  File config would win if available, then Env, then Flag, then Default\n")
	}

	// Example 5: Custom priority with specific sources
	fmt.Println("\n5. Custom Priority with Specific Sources:")
	cfg5 := commandkit.New()

	cfg5.Define("TIMEOUT").
		Int64().
		Env("TIMEOUT").
		Flag("timeout").
		Default(30).
		Sources(commandkit.SourceEnv, commandkit.SourceDefault).
		Priority(commandkit.SourcePriority{commandkit.SourceEnv, commandkit.SourceDefault})

	// Set environment variable
	os.Setenv("TIMEOUT", "60")

	result5 := cfg5.Process()
	if result5.Error != nil {
		fmt.Printf("Error: %v\n", result5.Error)
	} else {
		fmt.Printf("  TIMEOUT configured with only Env and Default sources\n")
		fmt.Printf("  Env value (60) would win over Default (30)\n")
		fmt.Printf("  Flag source is ignored in this configuration\n")
	}

	// Restore original args
	os.Args = originalArgs

	// Clean up environment variables
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("API_KEY")
	os.Unsetenv("TIMEOUT")

	fmt.Println("\n=== Priority System Features ===")
	fmt.Println("✅ Configurable source priority per definition")
	fmt.Println("✅ Config-level default priority")
	fmt.Println("✅ Preset priorities for common use cases")
	fmt.Println("✅ Override detection works with any priority order")
	fmt.Println("✅ Backward compatibility maintained")
	fmt.Println("✅ Comprehensive test coverage")

	fmt.Println("\n=== Available Priority Presets ===")
	fmt.Println("• PriorityFlagEnvDefault: Flag > Env > Default")
	fmt.Println("• PriorityEnvFlagDefault: Env > Flag > Default")
	fmt.Println("• PriorityFileEnvFlagDefault: File > Env > Flag > Default")
	fmt.Println("• PriorityDefaultOnly: Default only")

	fmt.Println("\n=== API Examples ===")
	fmt.Println("cfg.Define(\"PORT\").PriorityFlagEnvDefault()")
	fmt.Println("cfg.SetDefaultPriority(PriorityEnvFlagDefault)")
	fmt.Println("cfg.Define(\"KEY\").Priority([]SourceType{SourceEnv, SourceFlag})")
}
