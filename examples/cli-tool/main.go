// CLI Tool Example - Full-Featured Command Line Application
// Demonstrates: command system, middleware pipeline, authentication, help customization
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

	cfg.Define("TIMEOUT").
		Duration().
		Env("TIMEOUT").
		Flag("timeout").
		Default(30 * time.Second).
		Description("Operation timeout")

	cfg.Define("ADMIN_TOKEN").
		String().
		Env("ADMIN_TOKEN").
		Secret().
		Description("Admin authentication token")

	cfg.Define("API_KEY").
		String().
		Env("API_KEY").
		Secret().
		Description("API authentication key")

	// Set up global middleware
	setupGlobalMiddleware(cfg)

	// Set up commands
	setupCommands(cfg)

	// Execute with unified API
	if err := cfg.Execute(os.Args); err != nil {
		os.Exit(1)
	}
}

// setupGlobalMiddleware configures middleware that applies to all commands
func setupGlobalMiddleware(cfg *commandkit.Config) {
	// Recovery middleware - catches panics
	cfg.UseMiddleware(commandkit.RecoveryMiddleware())

	// Timing middleware - measures command execution time
	cfg.UseMiddleware(commandkit.TimingMiddleware())

	// Logging middleware - logs all command executions
	cfg.UseMiddleware(commandkit.DefaultLoggingMiddleware())

	// Error handling middleware - provides consistent error formatting
	cfg.UseMiddleware(commandkit.DefaultErrorHandlingMiddleware())

	// Metrics middleware - collects execution metrics
	cfg.UseMiddleware(commandkit.DefaultMetricsMiddleware())
}

// setupCommands defines all commands with their configurations and middleware
func setupCommands(cfg *commandkit.Config) {
	// Deploy command
	cfg.Command("deploy").
		Func(deployCommand).
		ShortHelp("Deploy application").
		LongHelp(`Deploy the application to the specified environment.

Usage: cli-tool deploy [options]

This command handles the complete deployment process including:
- Building the application
- Running tests
- Deploying to target environment
- Health checks`).
		Aliases("dep", "release").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("ENVIRONMENT").
				String().
				Flag("env").
				Required().
				OneOf("dev", "staging", "prod").
				Description("Target environment")

			cc.Define("DRY_RUN").
				Bool().
				Flag("dry-run").
				Default(false).
				Description("Perform a dry run without making changes")

			cc.Define("SKIP_TESTS").
				Bool().
				Flag("skip-tests").
				Default(false).
				Description("Skip running tests before deployment")

			cc.Define("FORCE").
				Bool().
				Flag("force").
				Default(false).
				Description("Force deployment even if checks fail")

			cc.Define("BRANCH").
				String().
				Flag("branch").
				Env("DEPLOY_BRANCH").
				Default("main").
				Description("Git branch to deploy")
		})

	// Add deploy-specific middleware
	cfg.UseMiddlewareForCommands([]string{"deploy"}, commandkit.RateLimitMiddleware(10, time.Minute))

	// Admin commands with authentication
	cfg.Command("admin-users").
		Func(adminUsersCommand).
		ShortHelp("Manage users").
		LongHelp(`Administrative command for user management.

Requires ADMIN_TOKEN environment variable for authentication.`).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("ACTION").
				String().
				Flag("action").
				Required().
				OneOf("list", "create", "delete", "update").
				Description("User action to perform")

			cc.Define("USERNAME").
				String().
				Flag("username").
				Description("Target username")

			cc.Define("ROLE").
				String().
				Flag("role").
				OneOf("admin", "user", "readonly").
				Default("user").
				Description("User role")
		})

	// Note: Authentication middleware would be added here in production

	cfg.Command("admin-shutdown").
		Func(adminShutdownCommand).
		ShortHelp("Shutdown the service").
		LongHelp(`Administrative command to shutdown the service.

Requires ADMIN_TOKEN environment variable for authentication.`).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("GRACEFUL").
				Bool().
				Flag("graceful").
				Default(true).
				Description("Perform graceful shutdown")

			cc.Define("DELAY").
				Duration().
				Flag("delay").
				Default(30 * time.Second).
				Description("Delay before shutdown")
		})

	// Status commands with API authentication
	cfg.Command("status").
		Func(statusCommand).
		ShortHelp("Show system status").
		LongHelp(`Show comprehensive system status including API and database.

Requires API_KEY environment variable for authentication.`).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("DETAILED").
				Bool().
				Flag("detailed").
				Default(false).
				Description("Show detailed status")

			cc.Define("FORMAT").
				String().
				Flag("format").
				Default("text").
				OneOf("text", "json").
				Description("Output format")
		})

	// Note: API authentication middleware would be added here in production

	cfg.Command("config").
		Func(configCommand).
		ShortHelp("Manage configuration").
		LongHelp(`Configuration management commands.

View, validate, and manage application configuration.`).
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("SHOW_SECRETS").
				Bool().
				Flag("show-secrets").
				Default(false).
				Description("Show secret values (use with caution)")

			cc.Define("VALIDATE_ONLY").
				Bool().
				Flag("validate-only").
				Default(false).
				Description("Only validate configuration")
		})

	// Help command with custom help
	cfg.Command("help").
		Func(helpCommand).
		ShortHelp("Show help").
		LongHelp(`Display comprehensive help for the CLI tool.

Use 'cli-tool help <command>' for command-specific help.`).
		Aliases("?", "--help", "-h")
}

// Command implementations

func deployCommand(ctx *commandkit.CommandContext) error {
	environment, err := commandkit.Get[string](ctx, "ENVIRONMENT")
	if err != nil {
		return fmt.Errorf("failed to get ENVIRONMENT: %w", err)
	}

	dryRun, err := commandkit.Get[bool](ctx, "DRY_RUN")
	if err != nil {
		return fmt.Errorf("failed to get DRY_RUN: %w", err)
	}

	skipTests, err := commandkit.Get[bool](ctx, "SKIP_TESTS")
	if err != nil {
		return fmt.Errorf("failed to get SKIP_TESTS: %w", err)
	}

	force, err := commandkit.Get[bool](ctx, "FORCE")
	if err != nil {
		return fmt.Errorf("failed to get FORCE: %w", err)
	}

	branch, err := commandkit.Get[string](ctx, "BRANCH")
	if err != nil {
		return fmt.Errorf("failed to get BRANCH: %w", err)
	}

	fmt.Printf("=== Deploy Command ===\n")
	fmt.Printf("Environment: %s\n", environment)
	fmt.Printf("Dry Run: %v\n", dryRun)
	fmt.Printf("Skip Tests: %v\n", skipTests)
	fmt.Printf("Force: %v\n", force)
	fmt.Printf("Branch: %s\n", branch)

	if dryRun {
		fmt.Printf("DRY RUN: No actual deployment performed\n")
	} else {
		fmt.Printf("Deployment would be performed here...\n")
	}

	return nil
}

func adminUsersCommand(ctx *commandkit.CommandContext) error {
	action, err := commandkit.Get[string](ctx, "ACTION")
	if err != nil {
		return fmt.Errorf("failed to get ACTION: %w", err)
	}

	username, err := commandkit.Get[string](ctx, "USERNAME")
	if err != nil {
		return fmt.Errorf("failed to get USERNAME: %w", err)
	}

	role, err := commandkit.Get[string](ctx, "ROLE")
	if err != nil {
		return fmt.Errorf("failed to get ROLE: %w", err)
	}

	fmt.Printf("=== Admin Users Command ===\n")
	fmt.Printf("Action: %s\n", action)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Role: %s\n", role)

	switch action {
	case "list":
		fmt.Printf("Users: admin, user1, user2\n")
	case "create":
		fmt.Printf("User %s created with role %s\n", username, role)
	case "delete":
		fmt.Printf("User %s deleted\n", username)
	case "update":
		fmt.Printf("User %s updated with role %s\n", username, role)
	}

	return nil
}

func adminShutdownCommand(ctx *commandkit.CommandContext) error {
	graceful, err := commandkit.Get[bool](ctx, "GRACEFUL")
	if err != nil {
		return fmt.Errorf("failed to get GRACEFUL: %w", err)
	}

	delay, err := commandkit.Get[time.Duration](ctx, "DELAY")
	if err != nil {
		return fmt.Errorf("failed to get DELAY: %w", err)
	}

	fmt.Printf("=== Admin Shutdown Command ===\n")
	fmt.Printf("Graceful: %v\n", graceful)
	fmt.Printf("Delay: %v\n", delay)

	if graceful {
		fmt.Printf("Initiating graceful shutdown in %v...\n", delay)
	} else {
		fmt.Printf("Initiating immediate shutdown...\n")
	}

	return nil
}

func statusCommand(ctx *commandkit.CommandContext) error {
	detailed, err := commandkit.Get[bool](ctx, "DETAILED")
	if err != nil {
		return fmt.Errorf("failed to get DETAILED: %w", err)
	}

	format, err := commandkit.Get[string](ctx, "FORMAT")
	if err != nil {
		return fmt.Errorf("failed to get FORMAT: %w", err)
	}

	fmt.Printf("=== Status Command ===\n")
	fmt.Printf("Detailed: %v\n", detailed)
	fmt.Printf("Format: %s\n", format)
	fmt.Printf("API Status: Healthy\n")
	fmt.Printf("Response Time: 45ms\n")
	fmt.Printf("Uptime: 7 days, 14 hours\n")

	if detailed {
		fmt.Printf("Database: Connected\n")
		fmt.Printf("Cache: Active\n")
		fmt.Printf("Queue: Normal\n")
	}

	return nil
}

func configCommand(ctx *commandkit.CommandContext) error {
	showSecrets, err := commandkit.Get[bool](ctx, "SHOW_SECRETS")
	if err != nil {
		return fmt.Errorf("failed to get SHOW_SECRETS: %w", err)
	}

	validateOnly, err := commandkit.Get[bool](ctx, "VALIDATE_ONLY")
	if err != nil {
		return fmt.Errorf("failed to get VALIDATE_ONLY: %w", err)
	}

	fmt.Printf("=== Config Command ===\n")
	fmt.Printf("Show Secrets: %v\n", showSecrets)
	fmt.Printf("Validate Only: %v\n", validateOnly)

	if !validateOnly {
		fmt.Printf("Current Configuration:\n")
		fmt.Printf("  LOG_LEVEL: info\n")
		fmt.Printf("  TIMEOUT: 30s\n")
		if showSecrets {
			fmt.Printf("  ADMIN_TOKEN: [REDACTED]\n")
			fmt.Printf("  API_KEY: [REDACTED]\n")
		}
	}

	return nil
}

func helpCommand(ctx *commandkit.CommandContext) error {
	fmt.Printf("=== Help Command ===\n")
	fmt.Printf("CLI Tool - Comprehensive Command Management\n\n")
	fmt.Printf("Available Commands:\n")
	fmt.Printf("  deploy        Deploy application\n")
	fmt.Printf("  admin-users   Manage users\n")
	fmt.Printf("  admin-shutdown Shutdown the service\n")
	fmt.Printf("  status        Show system status\n")
	fmt.Printf("  config        Manage configuration\n")
	fmt.Printf("  help          Show this help\n\n")
	fmt.Printf("Use 'cli-tool <command> --help' for command-specific help\n")
	return nil
}
