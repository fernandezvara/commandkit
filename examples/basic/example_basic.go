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

	// Process configuration
	if errs := cfg.Process(); len(errs) > 0 {
		cfg.PrintErrors(errs)
		os.Exit(1)
	}

	// Use configuration with type safety
	port := commandkit.Get[int64](cfg, "PORT")
	baseURL := commandkit.Get[string](cfg, "BASE_URL")
	corsOrigins := commandkit.Get[[]string](cfg, "CORS_ORIGINS")
	tokenTTL := commandkit.Get[time.Duration](cfg, "ACCESS_TOKEN_TTL")

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
