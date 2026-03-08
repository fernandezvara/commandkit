// Web Server Example - Complete Production Application
// Demonstrates: config-only mode, validation, secrets, file loading, source priority, builder patterns
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fernandezvara/commandkit"
)

func main() {
	cfg := commandkit.New()

	// Server configuration with comprehensive validation
	cfg.Define("PORT").
		Int64().
		Env("PORT").
		Flag("port").
		Default(int64(8080)).
		Range(1, 65535).
		Description("HTTP server port")

	cfg.Define("HOST").
		String().
		Env("HOST").
		Flag("host").
		Default("localhost").
		Description("Server host")

	cfg.Define("BASE_URL").
		String().
		Env("BASE_URL").
		Flag("base-url").
		Required().
		URL().
		Description("Public base URL of the service")

	cfg.Define("DATABASE_URL").
		String().
		Env("DATABASE_URL").
		Required().
		Secret().
		MinLength(10).
		Description("Database connection string")

	cfg.Define("REDIS_URL").
		String().
		Env("REDIS_URL").
		Secret().
		Description("Redis connection URL")

	cfg.Define("LOG_LEVEL").
		String().
		Env("LOG_LEVEL").
		Flag("log-level").
		Default("info").
		OneOf("debug", "info", "warn", "error").
		Description("Logging level")

	cfg.Define("ACCESS_TOKEN_TTL").
		Duration().
		Env("ACCESS_TOKEN_TTL").
		Default(15 * time.Minute).
		MinDuration(1 * time.Minute).
		MaxDuration(24 * time.Hour).
		Description("Access token lifetime")

	cfg.Define("CORS_ORIGINS").
		StringSlice().
		Env("CORS_ORIGINS").
		Flag("cors-origins").
		Default([]string{"http://localhost:3000"}).
		Delimiter(",").
		Description("Allowed CORS origins")

	cfg.Define("JWT_SIGNING_KEY").
		String().
		Env("JWT_SIGNING_KEY").
		Required().
		Secret().
		MinLength(32).
		Description("Secret key for signing access tokens")

	cfg.Define("ENVIRONMENT").
		String().
		Env("ENVIRONMENT").
		Flag("env").
		Default("development").
		OneOf("development", "staging", "production").
		Description("Application environment")

	cfg.Define("MAX_CONNECTIONS").
		Int64().
		Env("MAX_CONNECTIONS").
		Default(int64(100)).
		Range(1, 1000).
		Description("Maximum database connections")

	cfg.Define("ENABLE_METRICS").
		Bool().
		Env("ENABLE_METRICS").
		Flag("enable-metrics").
		Default(true).
		Description("Enable metrics collection")

	// Demonstrate builder pattern cloning for similar configurations
	baseTimeoutConfig := cfg.Define("READ_TIMEOUT").
		Duration().
		Env("READ_TIMEOUT").
		Default(30 * time.Second).
		MinDuration(1 * time.Second).
		MaxDuration(5 * time.Minute).
		Description("Read timeout")

	// Clone and customize for write timeout
	baseTimeoutConfig.Clone().
		Env("WRITE_TIMEOUT").
		Flag("write-timeout").
		Default(60 * time.Second).
		Description("Write timeout")

	// Clone and customize for idle timeout
	baseTimeoutConfig.Clone().
		Env("IDLE_TIMEOUT").
		Flag("idle-timeout").
		Default(120 * time.Second).
		Description("Idle timeout")

	// Set up file-based configuration based on environment
	if err := setupFileConfig(cfg); err != nil {
		fmt.Printf("Warning: Could not set up file config: %v\n", err)
	}

	// Execute configuration (single unified entry point)
	if err := cfg.Execute(os.Args); err != nil {
		os.Exit(1)
	}
	defer cfg.Destroy()

	// Create command context for accessing configuration
	ctx := commandkit.NewCommandContext([]string{}, cfg, "web-server", "")

	// Use configuration with type safety
	port, err := commandkit.Get[int64](ctx, "PORT")
	if err != nil {
		log.Fatalf("Error getting PORT: %v", err)
	}

	host, err := commandkit.Get[string](ctx, "HOST")
	if err != nil {
		log.Fatalf("Error getting HOST: %v", err)
	}

	baseURL, err := commandkit.Get[string](ctx, "BASE_URL")
	if err != nil {
		log.Fatalf("Error getting BASE_URL: %v", err)
	}

	logLevel, err := commandkit.Get[string](ctx, "LOG_LEVEL")
	if err != nil {
		log.Fatalf("Error getting LOG_LEVEL: %v", err)
	}

	environment, err := commandkit.Get[string](ctx, "ENVIRONMENT")
	if err != nil {
		log.Fatalf("Error getting ENVIRONMENT: %v", err)
	}

	// Access secrets safely
	dbURL := cfg.GetSecret("DATABASE_URL")
	if !dbURL.IsSet() {
		log.Fatal("Database URL is required")
	}

	redisURL := cfg.GetSecret("REDIS_URL")
	jwtKey := cfg.GetSecret("JWT_SIGNING_KEY")
	if !jwtKey.IsSet() {
		log.Fatal("JWT signing key is required")
	}

	// Display configuration summary
	fmt.Printf("=== Web Server Configuration ===\n")
	fmt.Printf("Environment: %s\n", environment)
	fmt.Printf("Server: %s:%d\n", host, port)
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Log Level: %s\n", logLevel)
	fmt.Printf("Database: configured (%d bytes)\n", dbURL.Size())

	if redisURL.IsSet() {
		fmt.Printf("Redis: configured (%d bytes)\n", redisURL.Size())
	} else {
		fmt.Printf("Redis: not configured\n")
	}

	fmt.Printf("JWT Key: configured (%d bytes)\n", jwtKey.Size())

	// Get optional configurations
	if enableMetrics, err := commandkit.Get[bool](ctx, "ENABLE_METRICS"); err == nil && enableMetrics {
		fmt.Printf("Metrics: enabled\n")
	}

	if maxConn, err := commandkit.Get[int64](ctx, "MAX_CONNECTIONS"); err == nil {
		fmt.Printf("Max DB Connections: %d\n", maxConn)
	}

	if corsOrigins, err := commandkit.Get[[]string](ctx, "CORS_ORIGINS"); err == nil {
		fmt.Printf("CORS Origins: %v\n", corsOrigins)
	}

	if tokenTTL, err := commandkit.Get[time.Duration](ctx, "ACCESS_TOKEN_TTL"); err == nil {
		fmt.Printf("Access Token TTL: %v\n", tokenTTL)
	}

	fmt.Printf("=== Server Starting ===\n")
	fmt.Printf("Server would start on %s:%d\n", host, port)
	fmt.Printf("Environment: %s\n", environment)
	fmt.Printf("All configuration validated successfully!\n")
}

// setupFileConfig configures file-based settings based on environment
func setupFileConfig(cfg *commandkit.Config) error {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	// Set config file based on environment
	configFile := fmt.Sprintf("config/%s.json", environment)
	if _, err := os.Stat(configFile); err == nil {
		// File configuration will be automatically loaded by priority system
		fmt.Printf("Will load configuration from: %s\n", configFile)
	} else {
		fmt.Printf("Config file not found: %s (using defaults)\n", configFile)
	}

	return nil
}
