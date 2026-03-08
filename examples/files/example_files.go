// Configuration files example demonstrating Story 27 completion
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fernandezvara/commandkit"
)

func main() {
	cfg := commandkit.New()

	// Server configuration
	cfg.Define("PORT").
		Int64().
		Default(8080).
		Range(1, 65535).
		Description("HTTP server port")

	cfg.Define("BASE_URL").
		String().
		Required().
		URL().
		Description("Public base URL of the service")

	cfg.Define("DEBUG").
		Bool().
		Default(false).
		Description("Enable debug mode")

	// Database configuration
	cfg.Define("DATABASE_URL").
		String().
		Required().
		Secret().
		Description("Database connection string")

	// JWT configuration
	cfg.Define("JWT_SIGNING_KEY").
		String().
		Required().
		Secret().
		MinLength(32).
		Description("JWT signing key")

	cfg.Define("ACCESS_TOKEN_TTL").
		Duration().
		Default(15 * time.Minute).
		Description("Access token lifetime")

	// Load configuration files
	fmt.Println("=== Loading Configuration Files ===")

	// Load base configuration
	err := cfg.LoadFile("config.yaml")
	if err != nil {
		fmt.Printf("Warning: Could not load config.yaml: %v\n", err)
	}

	// Load environment-specific overrides
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	err = cfg.SetEnvironment(env)
	if err != nil {
		fmt.Printf("Warning: Could not set environment: %v\n", err)
	}

	// Process configuration
	if result := cfg.Process(); result.Error != nil {
		fmt.Fprintln(os.Stderr, result.Message)
		os.Exit(1)
	}

	// Create command context for new API
	ctx := commandkit.NewCommandContext([]string{}, cfg, "example", "")

	// Use configuration
	port, err := commandkit.Get[int64](ctx, "PORT")
	if err != nil {
		fmt.Printf("Error getting PORT: %v\n", err)
		os.Exit(1)
	}
	baseURL, err := commandkit.Get[string](ctx, "BASE_URL")
	if err != nil {
		fmt.Printf("Error getting BASE_URL: %v\n", err)
		os.Exit(1)
	}
	debug, err := commandkit.Get[bool](ctx, "DEBUG")
	if err != nil {
		fmt.Printf("Error getting DEBUG: %v\n", err)
		os.Exit(1)
	}
	tokenTTL, err := commandkit.Get[time.Duration](ctx, "ACCESS_TOKEN_TTL")
	if err != nil {
		fmt.Printf("Error getting ACCESS_TOKEN_TTL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Configuration Loaded (Environment: %s) ===\n", env)
	fmt.Printf("Server starting on port %d\n", port)
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Debug mode: %v\n", debug)
	fmt.Printf("Token TTL: %s\n", tokenTTL)

	// Access secrets safely
	jwtKey := cfg.GetSecret("JWT_SIGNING_KEY")
	dbURL := cfg.GetSecret("DATABASE_URL")

	if jwtKey.IsSet() {
		fmt.Printf("JWT Key size: %d bytes\n", jwtKey.Size())
	}

	if dbURL.IsSet() {
		fmt.Printf("Database URL size: %d bytes\n", dbURL.Size())
	}

	// Show all configuration (secrets masked)
	fmt.Printf("\n=== Full Configuration ===\n")
	for k, v := range cfg.Dump() {
		fmt.Printf("  %s = %s\n", k, v)
	}

	// Clean up secrets
	defer cfg.Destroy()
}
