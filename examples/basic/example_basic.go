// Basic configuration example demonstrating Story 1 completion
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
		Env("PORT").
		Flag("port").
		Default(int64(8080)).
		Range(1, 65535).
		Description("HTTP server port")

	cfg.Define("BASE_URL").
		String().
		Env("BASE_URL").
		Flag("base-url").
		Required().
		URL().
		Description("Public base URL of the service")

	// Database configuration
	cfg.Define("DATABASE_URL").
		String().
		Env("DATABASE_URL").
		Required().
		Secret().
		MinLength(10).
		Description("PostgreSQL connection string")

	// JWT configuration
	cfg.Define("JWT_SIGNING_KEY").
		String().
		Env("JWT_SIGNING_KEY").
		Required().
		Secret().
		MinLength(32).
		Description("Secret key for signing access tokens")

	cfg.Define("ACCESS_TOKEN_TTL").
		Duration().
		Env("ACCESS_TOKEN_TTL").
		Default(15*time.Minute).
		DurationRange(1*time.Minute, 24*time.Hour).
		Description("Access token lifetime")

	// CORS configuration
	cfg.Define("CORS_ORIGINS").
		StringSlice().
		Env("CORS_ORIGINS").
		Flag("cors-origins").
		Delimiter(",").
		Default([]string{"http://localhost:3000"}).
		Description("Allowed CORS origins")

	if err := cfg.Execute(os.Args); err != nil {
		os.Exit(1)
	}

	// Use configuration with type safety
	// Create a command context for the new API
	ctx := commandkit.NewCommandContext([]string{}, cfg, "example", "")

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
	corsOrigins, err := commandkit.Get[[]string](ctx, "CORS_ORIGINS")
	if err != nil {
		fmt.Printf("Error getting CORS_ORIGINS: %v\n", err)
		os.Exit(1)
	}
	tokenTTL, err := commandkit.Get[time.Duration](ctx, "ACCESS_TOKEN_TTL")
	if err != nil {
		fmt.Printf("Error getting ACCESS_TOKEN_TTL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Configuration Loaded ===\n")
	fmt.Printf("Server starting on port %d\n", port)
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("CORS Origins: %v\n", corsOrigins)
	fmt.Printf("Token TTL: %s\n", tokenTTL)

	// Access secrets safely
	jwtKey := cfg.GetSecret("JWT_SIGNING_KEY")
	dbURL := cfg.GetSecret("DATABASE_URL")

	fmt.Printf("JWT Key size: %d bytes\n", jwtKey.Size())
	fmt.Printf("Database URL size: %d bytes\n", dbURL.Size())

	// Clean up secrets
	defer cfg.Destroy()

	// Show all configuration (secrets masked)
	fmt.Printf("\n=== Full Configuration ===\n")
	for k, v := range cfg.Dump() {
		fmt.Printf("  %s = %s\n", k, v)
	}
}
