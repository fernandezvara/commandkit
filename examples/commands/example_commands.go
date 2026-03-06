// Command system example demonstrating Stories 17-18 completion
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fernandezvara/commandkit"
)

func main() {
	cfg := commandkit.New()

	// Global configuration (available to all commands)
	cfg.Define("VERBOSE").
		Bool().
		Env("VERBOSE").
		Flag("verbose").
		Default(false).
		Description("Enable verbose logging")

	cfg.Define("LOG_LEVEL").
		String().
		Env("LOG_LEVEL").
		Flag("log-level").
		Default("info").
		OneOf("debug", "info", "warn", "error").
		Description("Logging level")

	// Start command with subcommands
	cfg.Command("start").
		Func(startCommand).
		ShortHelp("Start the service").
		LongHelp(`Start the service with all components initialized.
		
Usage: myapp start [options] [subcommand]

This will initialize the database, start HTTP server, and begin accepting connections.
For production use, consider using the --daemon flag.`).
		Aliases("s", "run", "up").
		Config(func(cc *commandkit.CommandConfig) {
			// Server configuration
			cc.Define("PORT").
				Int64().
				Flag("port").
				Default(int64(8080)).
				Range(1, 65535).
				Description("HTTP server port")

			cc.Define("BASE_URL").
				String().
				Flag("base-url").
				Required().
				URL().
				Description("Public base URL of the service")

			// Database configuration
			cc.Define("DATABASE_URL").
				String().
				Env("DATABASE_URL").
				Required().
				Secret().
				Description("Database connection string")

			// Daemon mode
			cc.Define("DAEMON").
				Bool().
				Flag("daemon").
				Default(false).
				Description("Run in background")
		}).
		SubCommand("server").
		Func(startServerCommand).
		ShortHelp("Start only the server").
		Aliases("srv").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("WORKERS").
				Int64Slice().
				Flag("workers").
				Delimiter(",").
				Default([]int64{1}).
				Description("Number of worker processes")
		}).
		SubCommand("worker").
		Func(startWorkerCommand).
		ShortHelp("Start only worker processes").
		Aliases("wrk").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("COUNT").
				Int64().
				Flag("count").
				Default(1).
				Range(1, 100).
				Description("Number of worker processes")
		})

	// Stop command
	cfg.Command("stop").
		Func(stopCommand).
		ShortHelp("Stop the service gracefully").
		LongHelp(`Stop the service gracefully with proper cleanup.
		
Usage: myapp stop [options]

This will attempt to gracefully shut down all components,
close database connections, and exit cleanly.`).
		Aliases("quit", "exit").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("TIMEOUT").
				Duration().
				Flag("timeout").
				Default(30*time.Second).
				DurationRange(5*time.Second, 5*time.Minute).
				Description("Graceful shutdown timeout")
		})

	// Status command
	cfg.Command("status").
		Func(statusCommand).
		ShortHelp("Show service status").
		LongHelp(`Display the current status of the service including:
- Running state
- Uptime
- Active connections
- Resource usage
- Configuration summary`)

	// Config command
	cfg.Command("config").
		Func(configCommand).
		ShortHelp("Show configuration").
		LongHelp(`Display current configuration values in a readable format.
Secrets are masked for security.`)

	// Execute commands
	if err := cfg.Execute(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func startCommand(ctx *commandkit.CommandContext) error {
	// Process command-specific configuration
	var config *commandkit.Config
	if ctx.CommandConfig != nil {
		config = ctx.CommandConfig
	} else {
		config = ctx.GlobalConfig
	}
	if result := config.Process(); result.Error != nil {
		if result.Message != "" {
			fmt.Fprintln(os.Stderr, result.Message)
		}
		return fmt.Errorf("configuration errors")
	}

	// Get configuration values
	portResult := commandkit.Get[int64](ctx, "PORT")
	if portResult.Error != nil {
		return fmt.Errorf("failed to get PORT: %w", portResult.Error)
	}
	port := commandkit.GetValue[int64](portResult)

	baseURLResult := commandkit.Get[string](ctx, "BASE_URL")
	if baseURLResult.Error != nil {
		return fmt.Errorf("failed to get BASE_URL: %w", baseURLResult.Error)
	}
	baseURL := commandkit.GetValue[string](baseURLResult)

	daemonResult := commandkit.Get[bool](ctx, "DAEMON")
	if daemonResult.Error != nil {
		return fmt.Errorf("failed to get DAEMON: %w", daemonResult.Error)
	}
	daemon := commandkit.GetValue[bool](daemonResult)

	verboseResult := commandkit.Get[bool](ctx, "VERBOSE")
	if verboseResult.Error != nil {
		return fmt.Errorf("failed to get VERBOSE: %w", verboseResult.Error)
	}
	verbose := commandkit.GetValue[bool](verboseResult)

	logLevelResult := commandkit.Get[string](ctx, "LOG_LEVEL")
	if logLevelResult.Error != nil {
		return fmt.Errorf("failed to get LOG_LEVEL: %w", logLevelResult.Error)
	}
	logLevel := commandkit.GetValue[string](logLevelResult)

	// Access secrets safely
	dbURL := config.GetSecret("DATABASE_URL")

	fmt.Printf("=== Starting Service ===\n")
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Daemon mode: %v\n", daemon)
	fmt.Printf("Verbose: %v\n", verbose)
	fmt.Printf("Log Level: %s\n", logLevel)

	if dbURL.IsSet() {
		fmt.Printf("Database: Connected (URL size: %d bytes)\n", dbURL.Size())
	}

	fmt.Printf("Service started successfully!\n")
	return nil
}

func startServerCommand(ctx *commandkit.CommandContext) error {
	// Process command-specific configuration
	var config *commandkit.Config
	if ctx.CommandConfig != nil {
		config = ctx.CommandConfig
	} else {
		config = ctx.GlobalConfig
	}
	if result := config.Process(); result.Error != nil {
		if result.Message != "" {
			fmt.Fprintln(os.Stderr, result.Message)
		}
		return fmt.Errorf("configuration errors")
	}

	workersResult := commandkit.Get[[]int64](ctx, "WORKERS")
	if workersResult.Error != nil {
		return fmt.Errorf("failed to get WORKERS: %w", workersResult.Error)
	}
	workers := commandkit.GetValue[[]int64](workersResult)

	portResult := commandkit.Get[int64](ctx, "PORT")
	if portResult.Error != nil {
		return fmt.Errorf("failed to get PORT: %w", portResult.Error)
	}
	port := commandkit.GetValue[int64](portResult)

	fmt.Printf("=== Starting Server ===\n")
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Workers: %v\n", workers)
	fmt.Printf("Server started successfully!\n")
	return nil
}

func startWorkerCommand(ctx *commandkit.CommandContext) error {
	// Process command-specific configuration
	var config *commandkit.Config
	if ctx.CommandConfig != nil {
		config = ctx.CommandConfig
	} else {
		config = ctx.GlobalConfig
	}
	if result := config.Process(); result.Error != nil {
		if result.Message != "" {
			fmt.Fprintln(os.Stderr, result.Message)
		}
		return fmt.Errorf("configuration errors")
	}

	countResult := commandkit.Get[int64](ctx, "COUNT")
	if countResult.Error != nil {
		return fmt.Errorf("failed to get COUNT: %w", countResult.Error)
	}
	count := commandkit.GetValue[int64](countResult)

	fmt.Printf("=== Starting Workers ===\n")
	fmt.Printf("Worker count: %d\n", count)
	fmt.Printf("Workers started successfully!\n")
	return nil
}

func stopCommand(ctx *commandkit.CommandContext) error {
	// Process command-specific configuration
	var config *commandkit.Config
	if ctx.CommandConfig != nil {
		config = ctx.CommandConfig
	} else {
		config = ctx.GlobalConfig
	}
	if result := config.Process(); result.Error != nil {
		if result.Message != "" {
			fmt.Fprintln(os.Stderr, result.Message)
		}
		return fmt.Errorf("configuration errors")
	}

	timeoutResult := commandkit.Get[time.Duration](ctx, "TIMEOUT")
	if timeoutResult.Error != nil {
		return fmt.Errorf("failed to get TIMEOUT: %w", timeoutResult.Error)
	}
	timeout := commandkit.GetValue[time.Duration](timeoutResult)

	fmt.Printf("=== Stopping Service ===\n")
	fmt.Printf("Graceful shutdown timeout: %v\n", timeout)
	fmt.Printf("Service stopped successfully!\n")
	return nil
}

func statusCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("=== Service Status ===\n")
	fmt.Printf("State: Running\n")
	fmt.Printf("Uptime: %v\n", time.Since(time.Now())) // This would be real uptime
	fmt.Printf("Active connections: 42\n")
	fmt.Printf("Memory usage: 128MB\n")
	fmt.Printf("CPU usage: 15%%\n")
	return nil
}

func configCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("=== Configuration ===\n")

	var config *commandkit.Config
	if ctx.CommandConfig != nil {
		config = ctx.CommandConfig
	} else {
		config = ctx.GlobalConfig
	}

	for k, v := range config.Dump() {
		fmt.Printf("  %s = %s\n", k, v)
	}

	return nil
}
