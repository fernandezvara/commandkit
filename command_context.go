// commandkit/command_context.go
package commandkit

// CommandContext provides context for command execution
type CommandContext struct {
	Args          []string
	GlobalConfig  *Config // Immutable global config
	CommandConfig *Config // Immutable command-specific config (nil if no command defs)
	Command       string
	SubCommand    string
	Flags         map[string]string
	data          map[string]any    // For middleware data sharing
	execution     *ExecutionContext // Thread-safe error collection
}

// NewCommandContext creates a new command context
func NewCommandContext(args []string, config *Config, command, subCommand string) *CommandContext {
	return &CommandContext{
		Args:          args,
		GlobalConfig:  config,
		CommandConfig: nil, // Will be set by ConfigProcessor if command has definitions
		Command:       command,
		SubCommand:    subCommand,
		Flags:         make(map[string]string),
		data:          make(map[string]any),
		execution:     NewExecutionContext(command), // Always initialize execution context
	}
}

// Set stores data in the context for middleware sharing
func (ctx *CommandContext) Set(key string, value any) {
	if ctx.data == nil {
		ctx.data = make(map[string]any)
	}
	ctx.data[key] = value
}

// GetData retrieves data from the context (renamed from Get to avoid naming conflict)
func (ctx *CommandContext) GetData(key string) (any, bool) {
	if ctx.data == nil {
		return nil, false
	}
	value, exists := ctx.data[key]
	return value, exists
}
