package commandkit

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTypesBasic(t *testing.T) {
	cfg := New()

	// Test unsigned integers
	cfg.Define("PORT").Uint().Flag("port").Default(8080)
	cfg.Define("BYTE").Uint8().Default(255)
	cfg.Define("SHORT").Uint16().Default(65535)
	cfg.Define("WORD").Uint32().Default(4294967295)
	cfg.Define("DWORD").Uint64().Default(uint64(18446744073709551615))

	// Test float32
	cfg.Define("RATIO").Float32().Default(0.5)

	// Test time
	cfg.Define("START_TIME").Time().Default(time.Now())

	// Test new slice types
	cfg.Define("FLOATS").Float64Slice().Default([]float64{1.1, 2.2, 3.3})
	cfg.Define("BOOLEANS").BoolSlice().Default([]bool{true, false, true})

	// Test FileMode
	cfg.Define("PERMS").FileMode().Default(0644)

	// Test IP
	cfg.Define("BIND_IP").IP().Default("127.0.0.1")

	// Test UUID
	cfg.Define("INSTANCE_ID").UUID().Default("550e8400-e29b-41d4-a716-446655440000")

	// Test Path
	cfg.Define("CONFIG_FILE").Path().Default("~/.config/app/config.yaml")

	// Execute config
	err := cfg.Execute([]string{"test"})
	if err != nil {
		t.Fatalf("Config execution failed: %v", err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test unsigned integer retrieval
	port, err := Get[uint](ctx, "PORT")
	if err != nil {
		t.Fatalf("Failed to get PORT: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected PORT=8080, got %d", port)
	}

	// Test FileMode retrieval as os.FileMode
	perms, err := Get[os.FileMode](ctx, "PERMS")
	if err != nil {
		t.Fatalf("Failed to get PERMS: %v", err)
	}
	if perms != 0644 {
		t.Errorf("Expected PERMS=0644, got %v", perms)
	}

	// Test time retrieval
	startTime, err := Get[time.Time](ctx, "START_TIME")
	if err != nil {
		t.Fatalf("Failed to get START_TIME: %v", err)
	}
	if startTime.IsZero() {
		t.Error("Expected non-zero time")
	}

	// Test slice types
	floats, err := Get[[]float64](ctx, "FLOATS")
	if err != nil {
		t.Fatalf("Failed to get FLOATS: %v", err)
	}
	if len(floats) != 3 {
		t.Errorf("Expected 3 floats, got %d", len(floats))
	}

	booleans, err := Get[[]bool](ctx, "BOOLEANS")
	if err != nil {
		t.Fatalf("Failed to get BOOLEANS: %v", err)
	}
	if len(booleans) != 3 {
		t.Errorf("Expected 3 booleans, got %d", len(booleans))
	}
}

func TestPathExpansion(t *testing.T) {
	cfg := New()

	// Test path expansion
	cfg.Define("HOME_CONFIG").Path().Default("~/.config/test.yaml")
	cfg.Define("RELATIVE_CONFIG").Path().Default("./config.yaml")
	cfg.Define("ENV_CONFIG").Path().Default("$HOME/config.yaml")

	err := cfg.Execute([]string{"test"})
	if err != nil {
		t.Fatalf("Config execution failed: %v", err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test home expansion
	homeConfig, err := Get[string](ctx, "HOME_CONFIG")
	if err != nil {
		t.Fatalf("Failed to get HOME_CONFIG: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	expected := homeDir + "/.config/test.yaml"
	if homeConfig != expected {
		t.Errorf("Expected HOME_CONFIG=%s, got %s", expected, homeConfig)
	}

	// Test relative path expansion
	relConfig, err := Get[string](ctx, "RELATIVE_CONFIG")
	if err != nil {
		t.Fatalf("Failed to get RELATIVE_CONFIG: %v", err)
	}

	// Should be absolute path
	if relConfig[0] != '/' {
		t.Errorf("Expected absolute path, got %s", relConfig)
	}

	// Test environment variable expansion
	envConfig, err := Get[string](ctx, "ENV_CONFIG")
	if err != nil {
		t.Fatalf("Failed to get ENV_CONFIG: %v", err)
	}

	expectedEnv := homeDir + "/config.yaml"
	if envConfig != expectedEnv {
		t.Errorf("Expected ENV_CONFIG=%s, got %s", expectedEnv, envConfig)
	}
}

func TestFileModeParsing(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected uint64
	}{
		{"standard octal", "0755", 0755},
		{"without leading zero", "755", 0755},
		{"with 0o prefix", "0o755", 0755},
		{"uppercase prefix", "0O755", 0755},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCfg := New()
			testCfg.Define("PERMS").FileMode().Default(tc.input)

			err := testCfg.Execute([]string{"test"})
			if err != nil {
				t.Fatalf("Config execution failed: %v", err)
			}

			ctx := NewCommandContext([]string{}, testCfg, "test", "")
			perms, err := Get[os.FileMode](ctx, "PERMS")
			if err != nil {
				t.Fatalf("Failed to get PERMS: %v", err)
			}

			if uint64(perms) != tc.expected {
				t.Errorf("Expected %d, got %d", tc.expected, uint64(perms))
			}
		})
	}
}

func TestIPValidation(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"valid IPv4", "127.0.0.1", false},
		{"valid IPv6", "::1", false},
		{"invalid IP", "not.an.ip", true},
		{"empty", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCfg := New()
			testCfg.Define("IP").IP().Default(tc.input)

			err := testCfg.Execute([]string{"test"})
			if err != nil {
				t.Fatalf("Config execution failed: %v", err)
			}

			ctx := NewCommandContext([]string{}, testCfg, "test", "")
			ip, err := Get[string](ctx, "IP")

			if tc.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Failed to get IP: %v", err)
				}

				if ip != tc.input {
					t.Errorf("Expected %s, got %s", tc.input, ip)
				}
			}
		})
	}
}

func TestUUIDValidation(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid UUID uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"invalid UUID", "not-a-uuid", true},
		{"wrong format", "550e8400e29b41d4a71644665544zzzz", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCfg := New()
			testCfg.Define("ID").UUID().Default(tc.input)

			err := testCfg.Execute([]string{"test"})
			if err != nil {
				t.Fatalf("Config execution failed: %v", err)
			}

			ctx := NewCommandContext([]string{}, testCfg, "test", "")
			id, err := Get[uuid.UUID](ctx, "ID")

			if tc.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Failed to get ID: %v", err)
				}

				expected := uuid.MustParse(tc.input)
				if id != expected {
					t.Errorf("Expected %s, got %s", expected, id)
				}
			}
		})
	}
}
