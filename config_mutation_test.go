// commandkit/config_mutation_test.go
package commandkit

import (
	"testing"
)

// TestConfigMutationFix verifies that the config mutation issue has been resolved
// This test ensures that:
// 1. GlobalConfig remains unchanged during command execution
// 2. CommandConfig is set independently without affecting GlobalConfig
// 3. Get[T]() function correctly prioritizes CommandConfig over GlobalConfig
func TestConfigMutationFix(t *testing.T) {
	// Create a global config with some definitions
	globalConfig := New()
	globalConfig.Define("GLOBAL_PORT").Int64().Default(3000)
	globalConfig.Define("GLOBAL_MODE").String().Default("global")

	if err := globalConfig.Execute([]string{"test"}); err != nil {
		t.Fatalf("Global config processing failed: %v", err)
	}

	// Create a command with its own definitions
	cmd := NewCommand("test")
	cmd.Definitions["PORT"] = &Definition{key: "PORT", valueType: TypeInt64, flag: "port", defaultValue: int64(8080)}
	cmd.Definitions["MODE"] = &Definition{key: "MODE", valueType: TypeString, flag: "mode", defaultValue: "command"}

	// Create context with global config
	ctx := NewCommandContext([]string{"--port", "9000", "--mode", "test"}, globalConfig, "test", "")

	// Store the original global config reference
	originalGlobalConfig := ctx.GlobalConfig

	// Process command configuration (this used to mutate ctx.Config)
	services := newCommandServices()
	processor := services.ConfigProcessor
	result := processor.ProcessCommandConfig(cmd, ctx)

	// Verify processing succeeded
	if result.Error != nil {
		t.Fatalf("ProcessCommandConfig failed: %v", result.Error)
	}

	// 1. Verify GlobalConfig is unchanged (same reference)
	if ctx.GlobalConfig != originalGlobalConfig {
		t.Error("GlobalConfig reference should not change during command execution")
	}

	// 2. Verify CommandConfig is set and different from GlobalConfig
	if ctx.CommandConfig == nil {
		t.Error("CommandConfig should be set after processing")
	}
	if ctx.CommandConfig == ctx.GlobalConfig {
		t.Error("CommandConfig should be different from GlobalConfig")
	}

	// 3. Verify Get[T]() prioritizes CommandConfig values
	port, err := Get[int64](ctx, "PORT")
	if err != nil {
		t.Fatalf("Failed to get PORT: %v", err)
	}
	if port != 9000 {
		t.Errorf("Expected PORT=9000 from CommandConfig, got %d", port)
	}

	// 4. Verify Get[T]() falls back to GlobalConfig for keys not in CommandConfig
	globalMode, err := Get[string](ctx, "GLOBAL_MODE")
	if err != nil {
		t.Fatalf("Failed to get GLOBAL_MODE: %v", err)
	}
	if globalMode != "global" {
		t.Errorf("Expected GLOBAL_MODE='global' from GlobalConfig, got %s", globalMode)
	}

	// 5. Verify GlobalConfig values are unchanged
	globalPort, err := Get[int64](ctx, "GLOBAL_PORT")
	if err != nil {
		t.Fatalf("Failed to get GLOBAL_PORT: %v", err)
	}
	if globalPort != 3000 {
		t.Errorf("Expected GLOBAL_PORT=3000 unchanged, got %d", globalPort)
	}

	// 6. Verify CommandConfig has command-specific definitions
	if !ctx.CommandConfig.Has("PORT") {
		t.Error("CommandConfig should have PORT definition")
	}
	if !ctx.CommandConfig.Has("MODE") {
		t.Error("CommandConfig should have MODE definition")
	}

	// 7. Verify GlobalConfig doesn't have command-specific definitions
	if ctx.GlobalConfig.Has("PORT") {
		t.Error("GlobalConfig should not have command-specific PORT definition")
	}
	if ctx.GlobalConfig.Has("MODE") {
		t.Error("GlobalConfig should not have command-specific MODE definition")
	}
}

// TestConfigIsolation verifies that multiple commands don't interfere with each other
func TestConfigIsolation(t *testing.T) {
	globalConfig := New()
	globalConfig.Define("SHARED").String().Default("shared")

	if err := globalConfig.Execute([]string{"test"}); err != nil {
		t.Fatalf("Global config processing failed: %v", err)
	}

	// Create two different commands
	cmd1 := NewCommand("cmd1")
	cmd1.Definitions["VALUE"] = &Definition{key: "VALUE", valueType: TypeInt64, flag: "value", defaultValue: int64(100)}

	cmd2 := NewCommand("cmd2")
	cmd2.Definitions["VALUE"] = &Definition{key: "VALUE", valueType: TypeInt64, flag: "value", defaultValue: int64(200)}

	// Process first command
	ctx1 := NewCommandContext([]string{"--value", "150"}, globalConfig, "cmd1", "")
	services := newCommandServices()
	processor := services.ConfigProcessor
	result1 := processor.ProcessCommandConfig(cmd1, ctx1)

	if result1.Error != nil {
		t.Fatalf("ProcessCommandConfig failed for cmd1: %v", result1.Error)
	}

	// Process second command
	ctx2 := NewCommandContext([]string{"--value", "250"}, globalConfig, "cmd2", "")
	result2 := processor.ProcessCommandConfig(cmd2, ctx2)

	if result2.Error != nil {
		t.Fatalf("ProcessCommandConfig failed for cmd2: %v", result2.Error)
	}

	// Verify isolation
	value1, err := Get[int64](ctx1, "VALUE")
	if err != nil {
		t.Fatalf("Failed to get VALUE from ctx1: %v", err)
	}
	if value1 != 150 {
		t.Errorf("Expected ctx1 VALUE=150, got %d", value1)
	}

	value2, err := Get[int64](ctx2, "VALUE")
	if err != nil {
		t.Fatalf("Failed to get VALUE from ctx2: %v", err)
	}
	if value2 != 250 {
		t.Errorf("Expected ctx2 VALUE=250, got %d", value2)
	}

	// Verify shared config is accessible from both
	shared1, err := Get[string](ctx1, "SHARED")
	if err != nil {
		t.Fatalf("Failed to get SHARED from ctx1: %v", err)
	}
	if shared1 != "shared" {
		t.Errorf("Expected ctx1 SHARED='shared', got %s", shared1)
	}

	shared2, err := Get[string](ctx2, "SHARED")
	if err != nil {
		t.Fatalf("Failed to get SHARED from ctx2: %v", err)
	}
	if shared2 != "shared" {
		t.Errorf("Expected ctx2 SHARED='shared', got %s", shared2)
	}

	// Verify global config is unchanged
	if ctx1.GlobalConfig != ctx2.GlobalConfig {
		t.Error("Both contexts should reference the same GlobalConfig")
	}
	if ctx1.GlobalConfig != globalConfig {
		t.Error("GlobalConfig reference should not change")
	}
}
