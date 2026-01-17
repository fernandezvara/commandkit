package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fernandezvara/commandkit"
)

func main() {
	// Create a new config instance
	cfg := commandkit.New()

	// Define global configuration
	cfg.Define("PORT").Int64().Env("PORT").Flag("port").Default(8080).Range(1, 65535)
	cfg.Define("LOG_LEVEL").String().Env("LOG_LEVEL").Flag("log-level").Default("info").OneOf("debug", "info", "warn", "error")
	cfg.Define("ADMIN_TOKEN").String().Env("ADMIN_TOKEN").Secret()
	cfg.Define("API_KEY").String().Env("API_KEY").Secret()

	// Define commands with middleware
	setupCommands(cfg)

	// Add global middleware that applies to all commands
	cfg.UseMiddleware(commandkit.RecoveryMiddleware())
	cfg.UseMiddleware(commandkit.TimingMiddleware())
	cfg.UseMiddleware(commandkit.DefaultLoggingMiddleware())
	cfg.UseMiddleware(commandkit.DefaultErrorHandlingMiddleware())
	cfg.UseMiddleware(commandkit.DefaultMetricsMiddleware())

	// Add authentication middleware only for admin commands
	cfg.UseMiddlewareForCommands([]string{"admin", "admin-users", "admin-shutdown"},
		commandkit.TokenAuthMiddleware("ADMIN_TOKEN"),
	)

	// Add API key authentication for API commands
	cfg.UseMiddlewareForCommands([]string{"api", "api-status"},
		commandkit.TokenAuthMiddleware("API_KEY"),
	)

	// Add rate limiting for status commands
	cfg.UseMiddlewareForCommands([]string{"status", "api-status"},
		commandkit.RateLimitMiddleware(5, time.Minute),
	)

	// Process configuration
	if errs := cfg.Process(); len(errs) > 0 {
		cfg.PrintErrors(errs)
		os.Exit(1)
	}

	// Execute commands
	if err := cfg.Execute(os.Args); err != nil {
		log.Printf("Command execution failed: %v", err)
		os.Exit(1)
	}
}

func setupCommands(cfg *commandkit.Config) {
	// Status command - shows system status
	cfg.Command("status").
		ShortHelp("Show system status").
		LongHelp("Displays detailed system status including uptime, memory usage, and active connections.").
		Func(statusCommand).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("VERBOSE").Bool().Flag("verbose").Default(false).
				Description("Enable verbose output")
		})

	// API status command - shows API-specific status
	cfg.Command("api").
		ShortHelp("API operations").
		Func(apiCommand).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("ENDPOINT").String().Env("API_ENDPOINT").Default("https://api.example.com").
				Description("API endpoint URL")
		}).
		SubCommand("status").
		ShortHelp("Show API status").
		Func(apiStatusCommand)

	// Admin commands
	cfg.Command("admin").
		ShortHelp("Administrative operations").
		Func(adminCommand).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("FORCE").Bool().Flag("force").Default(false).
				Description("Force operation without confirmation")
		}).
		SubCommand("users").
		ShortHelp("Manage users").
		Func(adminUsersCommand).
		SubCommand("shutdown").
		ShortHelp("Shutdown the server").
		Func(adminShutdownCommand)

	// Deploy command - demonstrates conditional middleware
	cfg.Command("deploy").
		ShortHelp("Deploy application").
		Func(deployCommand).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("ENVIRONMENT").String().Env("DEPLOY_ENV").Default("development").
				OneOf("development", "staging", "production").
				Description("Target environment for deployment")
			cc.Define("DRY_RUN").Bool().Flag("dry-run").Default(false).
				Description("Perform a dry run without making changes")
		})

	// Help command
	cfg.Command("help").
		ShortHelp("Show help").
		Func(helpCommand)
}

// Command implementations

func statusCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("📊 System Status\n")
	fmt.Printf("Command: %s\n", ctx.Command)

	// Get timing from middleware context
	if duration, exists := ctx.Get("duration"); exists {
		fmt.Printf("Execution time so far: %v\n", duration)
	}

	// Get verbose flag from command-specific config
	verbose := ctx.Config.GetBool("VERBOSE")
	if verbose {
		fmt.Printf("🔍 Verbose mode enabled\n")
		fmt.Printf("⏰ Current time: %s\n", time.Now().Format(time.RFC3339))
		fmt.Printf("🖥️  PID: %d\n", os.Getpid())
	}

	fmt.Printf("✅ Status check completed\n")
	return nil
}

func apiCommand(ctx *commandkit.CommandContext) error {
	endpoint := ctx.Config.GetString("ENDPOINT")
	fmt.Printf("🌐 API Operations\n")
	fmt.Printf("Endpoint: %s\n", endpoint)

	// Get auth token from middleware context
	if token, exists := ctx.Get("auth_token"); exists {
		fmt.Printf("🔑 Authenticated with token length: %d\n", len(token.(string)))
	}

	fmt.Printf("Use 'api status' for detailed API status\n")
	return nil
}

func apiStatusCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("🌐 API Status\n")

	// Get execution count from rate limiting middleware
	if count, exists := ctx.Get("execution_count"); exists {
		fmt.Printf("📊 Execution count: %d\n", count)
	}

	// Simulate API status check
	time.Sleep(100 * time.Millisecond) // Simulate API call

	fmt.Printf("✅ API is healthy\n")
	fmt.Printf("📈 Response time: 42ms\n")
	fmt.Printf("🔗 Connections: 127\n")

	return nil
}

func adminCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("🔐 Admin Operations\n")

	// Get auth token from middleware context
	if _, exists := ctx.Get("auth_token"); exists {
		fmt.Printf("🔑 Authenticated as admin\n")
	}

	force := ctx.Config.GetBool("FORCE")
	if force {
		fmt.Printf("⚠️  Force mode enabled\n")
	}

	fmt.Printf("Available subcommands: users, shutdown\n")
	return nil
}

func adminUsersCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("👥 User Management\n")

	// Simulate user management operations
	fmt.Printf("📋 Active users: 1,234\n")
	fmt.Printf("🔒 Locked users: 3\n")
	fmt.Printf("📝 Recent signups: 42\n")

	return nil
}

func adminShutdownCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("🛑 Shutdown Initiated\n")

	force := ctx.Config.GetBool("FORCE")
	if !force {
		fmt.Printf("⚠️  Use --force to confirm shutdown\n")
		return fmt.Errorf("shutdown requires confirmation")
	}

	fmt.Printf("🔄 Shutting down gracefully...\n")
	time.Sleep(1 * time.Second) // Simulate shutdown process
	fmt.Printf("✅ Shutdown complete\n")

	return nil
}

func deployCommand(ctx *commandkit.CommandContext) error {
	environment := ctx.Config.GetString("ENVIRONMENT")
	dryRun := ctx.Config.GetBool("DRY_RUN")

	fmt.Printf("🚀 Deployment\n")
	fmt.Printf("Environment: %s\n", environment)

	if dryRun {
		fmt.Printf("🔍 Dry run mode - no changes will be made\n")
	}

	// Simulate deployment steps
	steps := []string{"Building", "Testing", "Deploying", "Verifying"}
	for _, step := range steps {
		fmt.Printf("📦 %s...", step)
		time.Sleep(200 * time.Millisecond)
		fmt.Printf(" ✅\n")
	}

	if dryRun {
		fmt.Printf("🔍 Dry run completed successfully\n")
	} else {
		fmt.Printf("✅ Deployment to %s completed\n", environment)
	}

	return nil
}

func helpCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("📚 Command Help\n\n")

	fmt.Printf("Available commands:\n")
	fmt.Printf("  status     - Show system status\n")
	fmt.Printf("  api        - API operations\n")
	fmt.Printf("  admin      - Administrative operations\n")
	fmt.Printf("  deploy     - Deploy application\n")
	fmt.Printf("  help       - Show this help\n\n")

	fmt.Printf("Middleware in use:\n")
	fmt.Printf("  ✅ Recovery - Prevents crashes from panics\n")
	fmt.Printf("  ✅ Timing   - Measures execution time\n")
	fmt.Printf("  ✅ Logging  - Logs all command executions\n")
	fmt.Printf("  ✅ Error    - Handles errors consistently\n")
	fmt.Printf("  ✅ Metrics  - Collects command metrics\n")
	fmt.Printf("  🔐 Auth     - Authentication for admin/api commands\n")
	fmt.Printf("  📊 RateLimit - Rate limiting for status commands\n")

	return nil
}
