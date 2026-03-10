package commandkit

import (
	"testing"
	"time"
)

func TestDefinitionBuilderClone(t *testing.T) {
	cfg := New()

	// Create original builder
	original := cfg.Define("PORT").
		Int64().
		Env("PORT").
		Flag("port").
		Default(8080).
		Range(1, 65535).
		Required().
		Description("HTTP server port")

	// Clone the builder
	cloned := original.Clone()

	// Verify they are independent
	if original.def == cloned.def {
		t.Error("Cloned definition should be a different instance")
	}

	// Verify values are copied correctly
	if cloned.def.key != original.def.key {
		t.Errorf("Expected key %s, got %s", original.def.key, cloned.def.key)
	}
	if cloned.def.valueType != original.def.valueType {
		t.Errorf("Expected valueType %v, got %v", original.def.valueType, cloned.def.valueType)
	}
	if cloned.def.envVar != original.def.envVar {
		t.Errorf("Expected envVar %s, got %s", original.def.envVar, cloned.def.envVar)
	}
	if cloned.def.flag != original.def.flag {
		t.Errorf("Expected flag %s, got %s", original.def.flag, cloned.def.flag)
	}
	if cloned.def.defaultValue != original.def.defaultValue {
		t.Errorf("Expected defaultValue %v, got %v", original.def.defaultValue, cloned.def.defaultValue)
	}
	if cloned.def.required != original.def.required {
		t.Errorf("Expected required %v, got %v", original.def.required, cloned.def.required)
	}
	if cloned.def.description != original.def.description {
		t.Errorf("Expected description %s, got %s", original.def.description, cloned.def.description)
	}

	// Verify validations are copied
	if len(cloned.def.validations) != len(original.def.validations) {
		t.Errorf("Expected %d validations, got %d", len(original.def.validations), len(cloned.def.validations))
	}

	// Modify cloned builder and verify original is unaffected
	cloned.String().Default("default").Secret()
	if original.def.valueType == TypeString {
		t.Error("Original builder should not be affected by clone modifications")
	}
	if original.def.secret {
		t.Error("Original builder should not be affected by clone modifications")
	}
}

func TestCommandBuilderClone(t *testing.T) {
	cfg := New()

	// Create original command builder
	original := cfg.Command("start").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Start the service").
		LongHelp("Start the service with all components").
		Aliases("s", "run").
		Config(func(cc *CommandConfig) {
			cc.Define("PORT").Int64().Default(8080)
			cc.Define("HOST").String().Default("localhost")
		}).
		Middleware(func(next CommandFunc) CommandFunc { return next })

	// Clone the builder
	cloned := original.Clone()

	// Verify they are independent
	if original.cmd == cloned.cmd {
		t.Error("Cloned command should be a different instance")
	}

	// Verify values are copied correctly
	if cloned.cmd.Name != original.cmd.Name {
		t.Errorf("Expected name %s, got %s", original.cmd.Name, cloned.cmd.Name)
	}
	if cloned.cmd.ShortHelp != original.cmd.ShortHelp {
		t.Errorf("Expected ShortHelp %s, got %s", original.cmd.ShortHelp, cloned.cmd.ShortHelp)
	}
	if cloned.cmd.LongHelp != original.cmd.LongHelp {
		t.Errorf("Expected LongHelp %s, got %s", original.cmd.LongHelp, cloned.cmd.LongHelp)
	}

	// Verify aliases are copied
	if len(cloned.cmd.Aliases) != len(original.cmd.Aliases) {
		t.Errorf("Expected %d aliases, got %d", len(original.cmd.Aliases), len(cloned.cmd.Aliases))
	}
	for i, alias := range original.cmd.Aliases {
		if cloned.cmd.Aliases[i] != alias {
			t.Errorf("Expected alias %s at index %d, got %s", alias, i, cloned.cmd.Aliases[i])
		}
	}

	// Verify definitions are copied
	if len(cloned.cmd.Definitions) != len(original.cmd.Definitions) {
		t.Errorf("Expected %d definitions, got %d", len(original.cmd.Definitions), len(cloned.cmd.Definitions))
	}
	for key, def := range original.cmd.Definitions {
		clonedDef, exists := cloned.cmd.Definitions[key]
		if !exists {
			t.Errorf("Cloned command missing definition for key %s", key)
		}
		if clonedDef == def {
			t.Errorf("Cloned definition for key %s should be a different instance", key)
		}
	}

	// Verify middleware is copied
	if len(cloned.cmd.Middleware) != len(original.cmd.Middleware) {
		t.Errorf("Expected %d middleware, got %d", len(original.cmd.Middleware), len(cloned.cmd.Middleware))
	}

	// Modify cloned builder and verify original is unaffected
	cloned.ShortHelp("Modified help")
	if original.cmd.ShortHelp == "Modified help" {
		t.Error("Original command should not be affected by clone modifications")
	}
}

func TestBuilderCloneWithSubCommands(t *testing.T) {
	cfg := New()

	// Create command with subcommands
	original := cfg.Command("start").
		ShortHelp("Start the service").
		Config(func(cc *CommandConfig) {
			cc.Define("PORT").Int64().Default(8080)
		})

	// Add subcommand
	original.SubCommand("server").
		ShortHelp("Start server only").
		Config(func(cc *CommandConfig) {
			cc.Define("WORKERS").Int64().Default(4)
		})

	// Clone the builder
	cloned := original.Clone()

	// Verify subcommands are copied
	if len(cloned.cmd.SubCommands) != len(original.cmd.SubCommands) {
		t.Errorf("Expected %d subcommands, got %d", len(original.cmd.SubCommands), len(cloned.cmd.SubCommands))
	}

	clonedSubCmd, exists := cloned.cmd.SubCommands["server"]
	if !exists {
		t.Error("Cloned command missing subcommand 'server'")
	}

	if clonedSubCmd == original.cmd.SubCommands["server"] {
		t.Error("Cloned subcommand should be a different instance")
	}

	if clonedSubCmd.ShortHelp != original.cmd.SubCommands["server"].ShortHelp {
		t.Errorf("Expected subcommand ShortHelp %s, got %s",
			original.cmd.SubCommands["server"].ShortHelp, clonedSubCmd.ShortHelp)
	}

	// Verify subcommand definitions are copied
	if len(clonedSubCmd.Definitions) != len(original.cmd.SubCommands["server"].Definitions) {
		t.Errorf("Expected %d subcommand definitions, got %d",
			len(original.cmd.SubCommands["server"].Definitions), len(clonedSubCmd.Definitions))
	}
}

func TestConfigMethodSimplification(t *testing.T) {
	cfg := New()

	// Test that Config method still works correctly
	builder := cfg.Command("test").
		ShortHelp("Test command").
		Config(func(cc *CommandConfig) {
			cc.Define("PORT").Int64().Default(8080)
			cc.Define("HOST").String().Default("localhost")
		})

	// Verify definitions were added
	if len(builder.cmd.Definitions) != 2 {
		t.Errorf("Expected 2 definitions, got %d", len(builder.cmd.Definitions))
	}

	portDef, exists := builder.cmd.Definitions["PORT"]
	if !exists {
		t.Error("PORT definition not found")
	}
	if portDef.valueType != TypeInt64 {
		t.Errorf("Expected PORT to be Int64, got %v", portDef.valueType)
	}

	hostDef, exists := builder.cmd.Definitions["HOST"]
	if !exists {
		t.Error("HOST definition not found")
	}
	if hostDef.valueType != TypeString {
		t.Errorf("Expected HOST to be String, got %v", hostDef.valueType)
	}
}

func TestBuilderCloneForDRYPatterns(t *testing.T) {
	cfg := New()

	// Create a base configuration template
	baseConfig := func(cc *CommandConfig) {
		cc.Define("PORT").Int64().Flag("port").Default(8080).Range(1, 65535)
		cc.Define("HOST").String().Flag("host").Default("localhost")
		cc.Define("VERBOSE").Bool().Flag("verbose").Default(false)
	}

	// Create multiple commands with base configuration
	startCmd := cfg.Command("start").
		ShortHelp("Start the service").
		Config(func(cc *CommandConfig) {
			baseConfig(cc)
			cc.Define("WORKERS").Int64().Flag("workers").Default(4)
		})

	stopCmd := cfg.Command("stop").
		ShortHelp("Stop the service").
		Config(func(cc *CommandConfig) {
			baseConfig(cc)
			cc.Define("TIMEOUT").Duration().Flag("timeout").Default(30 * time.Second)
		})

	// Verify both commands have base configuration
	for _, cmd := range []*CommandBuilder{startCmd, stopCmd} {
		if len(cmd.cmd.Definitions) < 3 {
			t.Errorf("Expected at least 3 base definitions, got %d", len(cmd.cmd.Definitions))
		}

		// Check base definitions exist
		baseKeys := []string{"PORT", "HOST", "VERBOSE"}
		for _, key := range baseKeys {
			if _, exists := cmd.cmd.Definitions[key]; !exists {
				t.Errorf("Base definition %s not found in command", key)
			}
		}
	}

	// Verify command-specific definitions
	if _, exists := startCmd.cmd.Definitions["WORKERS"]; !exists {
		t.Error("WORKERS definition not found in start command")
	}

	if _, exists := stopCmd.cmd.Definitions["TIMEOUT"]; !exists {
		t.Error("TIMEOUT definition not found in stop command")
	}
}

func TestDefinitionCloneWithValidations(t *testing.T) {
	cfg := New()

	// Create definition with multiple validations
	original := cfg.Define("AGE").
		Int64().
		Min(18).
		Max(120).
		Required().
		Description("User age")

	// Clone and verify validations are copied
	cloned := original.Clone()

	if len(cloned.def.validations) != len(original.def.validations) {
		t.Errorf("Expected %d validations, got %d", len(original.def.validations), len(cloned.def.validations))
	}

	// Verify validation names are preserved
	originalNames := make(map[string]bool)
	for _, v := range original.def.validations {
		originalNames[v.Name] = true
	}

	for _, v := range cloned.def.validations {
		if !originalNames[v.Name] {
			t.Errorf("Unexpected validation %s in clone", v.Name)
		}
	}
}
