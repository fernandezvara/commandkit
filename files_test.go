package commandkit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigurationFiles(t *testing.T) {
	// Create test config files
	yamlConfig := `
port: 3000
base_url: "https://test.example.com"
debug: true
timeout: "5m"
cors_origins:
  - "https://test.example.com"
  - "https://api.example.com"
environments:
  development:
    port: 3001
    debug: true
  production:
    port: 80
    debug: false
`

	jsonConfig := `
{
  "database_url": "postgresql://test:test@localhost/testdb",
  "jwt_signing_key": "test-secret-key-that-is-32-chars-long",
  "rate_limit_rps": 1000
}
`

	tomlConfig := `
access_token_ttl = "30m"
refresh_token_ttl = "7d"
max_connections = 100
`

	// Write test files
	err := os.WriteFile("test.yaml", []byte(yamlConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}
	defer os.Remove("test.yaml")

	err = os.WriteFile("test.json", []byte(jsonConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}
	defer os.Remove("test.json")

	err = os.WriteFile("test.toml", []byte(tomlConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test TOML file: %v", err)
	}
	defer os.Remove("test.toml")

	cfg := New()

	// Define configuration
	cfg.Define("PORT").
		Int64().
		Default(8080).
		Range(1, 65535)

	cfg.Define("BASE_URL").
		String().
		Required().
		URL()

	cfg.Define("DEBUG").
		Bool().
		Default(false)

	cfg.Define("TIMEOUT").
		Duration().
		Default(30 * time.Second)

	cfg.Define("CORS_ORIGINS").
		StringSlice().
		Delimiter(",").
		Default([]string{})

	cfg.Define("DATABASE_URL").
		String().
		Secret()

	cfg.Define("JWT_SIGNING_KEY").
		String().
		Secret()

	cfg.Define("RATE_LIMIT_RPS").
		Float64().
		Default(100.0)

	cfg.Define("ACCESS_TOKEN_TTL").
		Duration().
		Default(15 * time.Minute)

	cfg.Define("REFRESH_TOKEN_TTL").
		Duration().
		Default(7 * 24 * time.Hour)

	cfg.Define("MAX_CONNECTIONS").
		Int64().
		Default(10)

	// Load configuration files
	err = cfg.LoadFiles("test.yaml", "test.json", "test.toml")
	if err != nil {
		t.Fatalf("Failed to load config files: %v", err)
	}

	// Process configuration
	errs := cfg.Process()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Logf("Config error: %s", err.Message)
		}
		t.Fatalf("Configuration errors: %v", errs)
	}

	// Test values from YAML
	port := Get[int64](cfg, "PORT")
	if port != 3000 {
		t.Errorf("Expected PORT=3000 from YAML, got %d", port)
	}

	baseURL := Get[string](cfg, "BASE_URL")
	if baseURL != "https://test.example.com" {
		t.Errorf("Expected BASE_URL from YAML, got %s", baseURL)
	}

	debug := Get[bool](cfg, "DEBUG")
	if !debug {
		t.Errorf("Expected DEBUG=true from YAML, got %v", debug)
	}

	timeout := Get[time.Duration](cfg, "TIMEOUT")
	if timeout != 5*time.Minute {
		t.Errorf("Expected TIMEOUT=5m from YAML, got %v", timeout)
	}

	corsOrigins := Get[[]string](cfg, "CORS_ORIGINS")
	expectedCORS := []string{"https://test.example.com", "https://api.example.com"}
	if len(corsOrigins) != len(expectedCORS) {
		t.Errorf("Expected %d CORS origins, got %d", len(expectedCORS), len(corsOrigins))
	}

	// Test values from JSON
	dbURL := cfg.GetSecret("DATABASE_URL")
	if !dbURL.IsSet() || dbURL.String() != "postgresql://test:test@localhost/testdb" {
		t.Error("DATABASE_URL not loaded correctly from JSON")
	}

	jwtKey := cfg.GetSecret("JWT_SIGNING_KEY")
	if !jwtKey.IsSet() || jwtKey.String() != "test-secret-key-that-is-32-chars-long" {
		t.Error("JWT_SIGNING_KEY not loaded correctly from JSON")
	}

	rateLimit := Get[float64](cfg, "RATE_LIMIT_RPS")
	if rateLimit != 1000.0 {
		t.Errorf("Expected RATE_LIMIT_RPS=1000 from JSON, got %f", rateLimit)
	}

	// Test values from TOML
	accessTokenTTL := Get[time.Duration](cfg, "ACCESS_TOKEN_TTL")
	if accessTokenTTL != 30*time.Minute {
		t.Errorf("Expected ACCESS_TOKEN_TTL=30m from TOML, got %v", accessTokenTTL)
	}

	refreshTokenTTL := Get[time.Duration](cfg, "REFRESH_TOKEN_TTL")
	if refreshTokenTTL != 7*24*time.Hour {
		t.Errorf("Expected REFRESH_TOKEN_TTL=7d from TOML, got %v", refreshTokenTTL)
	}

	maxConnections := Get[int64](cfg, "MAX_CONNECTIONS")
	if maxConnections != 100 {
		t.Errorf("Expected MAX_CONNECTIONS=100 from TOML, got %d", maxConnections)
	}
}

func TestEnvironmentSpecificConfiguration(t *testing.T) {
	yamlConfig := `
port: 8080
debug: false
environments:
  development:
    port: 3000
    debug: true
  production:
    port: 80
    debug: false
`

	err := os.WriteFile("test_env.yaml", []byte(yamlConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test environment YAML file: %v", err)
	}
	defer os.Remove("test_env.yaml")

	cfg := New()

	cfg.Define("PORT").Int64().Default(8080)
	cfg.Define("DEBUG").Bool().Default(false)

	// Load configuration
	err = cfg.LoadFile("test_env.yaml")
	if err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	// Test without environment
	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	port := Get[int64](cfg, "PORT")
	if port != 8080 {
		t.Errorf("Expected PORT=8080 (base), got %d", port)
	}

	debug := Get[bool](cfg, "DEBUG")
	if debug {
		t.Errorf("Expected DEBUG=false (base), got %v", debug)
	}

	// Test with development environment
	err = cfg.SetEnvironment("development")
	if err != nil {
		t.Fatalf("Failed to set environment: %v", err)
	}

	// Re-process with environment
	errs = cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	port = Get[int64](cfg, "PORT")
	if port != 3000 {
		t.Errorf("Expected PORT=3000 (development), got %d", port)
	}

	debug = Get[bool](cfg, "DEBUG")
	if !debug {
		t.Errorf("Expected DEBUG=true (development), got %v", debug)
	}

	// Test with production environment
	err = cfg.SetEnvironment("production")
	if err != nil {
		t.Fatalf("Failed to set environment: %v", err)
	}

	// Re-process with environment
	errs = cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	port = Get[int64](cfg, "PORT")
	if port != 80 {
		t.Errorf("Expected PORT=80 (production), got %d", port)
	}

	debug = Get[bool](cfg, "DEBUG")
	if debug {
		t.Errorf("Expected DEBUG=false (production), got %v", debug)
	}
}

func TestFilePriority(t *testing.T) {
	// Create two config files where second overrides first
	config1 := `
port: 3000
debug: true
timeout: "5m"
`

	config2 := `
port: 8080
timeout: "10m"
log_level: "info"
`

	err := os.WriteFile("test1.yaml", []byte(config1), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config1 file: %v", err)
	}
	defer os.Remove("test1.yaml")

	err = os.WriteFile("test2.yaml", []byte(config2), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config2 file: %v", err)
	}
	defer os.Remove("test2.yaml")

	cfg := New()

	cfg.Define("PORT").Int64().Default(8080)
	cfg.Define("DEBUG").Bool().Default(false)
	cfg.Define("TIMEOUT").Duration().Default(30 * time.Second)
	cfg.Define("LOG_LEVEL").String().Default("info")

	// Load files in order (second overrides first)
	err = cfg.LoadFiles("test1.yaml", "test2.yaml")
	if err != nil {
		t.Fatalf("Failed to load config files: %v", err)
	}

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	// PORT should be from second file (override)
	port := Get[int64](cfg, "PORT")
	if port != 8080 {
		t.Errorf("Expected PORT=8080 (from second file), got %d", port)
	}

	// DEBUG should be from first file (not overridden)
	debug := Get[bool](cfg, "DEBUG")
	if !debug {
		t.Errorf("Expected DEBUG=true (from first file), got %v", debug)
	}

	// TIMEOUT should be from second file (override)
	timeout := Get[time.Duration](cfg, "TIMEOUT")
	if timeout != 10*time.Minute {
		t.Errorf("Expected TIMEOUT=10m (from second file), got %v", timeout)
	}

	// LOG_LEVEL should be from second file (only in second file)
	logLevel := Get[string](cfg, "LOG_LEVEL")
	if logLevel != "info" {
		t.Errorf("Expected LOG_LEVEL=info (from second file), got %s", logLevel)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Create a test config file
	yamlConfig := `
port: 9000
debug: true
`
	tmpFile := filepath.Join(os.TempDir(), "test_config_env.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Set environment variable pointing to config file
	os.Setenv("TEST_CONFIG_PATH", tmpFile)
	defer os.Unsetenv("TEST_CONFIG_PATH")

	cfg := New()
	cfg.Define("PORT").Int64().Default(8080)
	cfg.Define("DEBUG").Bool().Default(false)

	// Load from environment variable
	err = cfg.LoadFromEnv("TEST_CONFIG_PATH")
	if err != nil {
		t.Fatalf("LoadFromEnv failed: %v", err)
	}

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	port := Get[int64](cfg, "PORT")
	if port != 9000 {
		t.Errorf("Expected PORT=9000 from file, got %d", port)
	}

	debug := Get[bool](cfg, "DEBUG")
	if !debug {
		t.Errorf("Expected DEBUG=true from file, got %v", debug)
	}
}

func TestLoadFromEnvNotSet(t *testing.T) {
	cfg := New()

	// LoadFromEnv returns nil when env var is not set (no-op)
	err := cfg.LoadFromEnv("NONEXISTENT_CONFIG_PATH")
	if err != nil {
		t.Errorf("LoadFromEnv should return nil for unset env var, got: %v", err)
	}
}

func TestSetEnvironmentFromEnv(t *testing.T) {
	yamlConfig := `
port: 8080
environments:
  staging:
    port: 8081
  production:
    port: 80
`
	tmpFile := filepath.Join(os.TempDir(), "test_env_from_env.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Set environment variable for environment name
	os.Setenv("APP_ENV", "staging")
	defer os.Unsetenv("APP_ENV")

	cfg := New()
	cfg.Define("PORT").Int64().Default(8080)

	err = cfg.LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// Set environment from env var
	err = cfg.SetEnvironmentFromEnv("APP_ENV")
	if err != nil {
		t.Fatalf("SetEnvironmentFromEnv failed: %v", err)
	}

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	port := Get[int64](cfg, "PORT")
	if port != 8081 {
		t.Errorf("Expected PORT=8081 (staging), got %d", port)
	}
}

func TestSetEnvironmentFromEnvNotSet(t *testing.T) {
	cfg := New()

	// SetEnvironmentFromEnv returns nil when env var is not set (no-op)
	err := cfg.SetEnvironmentFromEnv("NONEXISTENT_ENV_VAR")
	if err != nil {
		t.Errorf("SetEnvironmentFromEnv should return nil for unset env var, got: %v", err)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	cfg := New()

	err := cfg.LoadFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadFile should return error for non-existent file")
	}
}

func TestLoadFileUnsupportedFormat(t *testing.T) {
	// Create a file with unsupported extension
	tmpFile := filepath.Join(os.TempDir(), "test_config.txt")
	err := os.WriteFile(tmpFile, []byte("some content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg := New()
	err = cfg.LoadFile(tmpFile)
	if err == nil {
		t.Error("LoadFile should return error for unsupported file format")
	}
}
