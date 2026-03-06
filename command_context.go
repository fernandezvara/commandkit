// commandkit/command_context.go
package commandkit

// CommandContext provides context for command execution
type CommandContext struct {
	Args       []string
	Config     *Config
	Command    string
	SubCommand string
	Flags      map[string]string
	data       map[string]any // For middleware data sharing
}

// NewCommandContext creates a new command context
func NewCommandContext(args []string, config *Config, command, subCommand string) *CommandContext {
	return &CommandContext{
		Args:       args,
		Config:     config,
		Command:    command,
		SubCommand: subCommand,
		Flags:      make(map[string]string),
		data:       make(map[string]any),
	}
}

// ContextGet retrieves a typed value from the context data using generics
func ContextGet[T any](ctx *CommandContext, key string) T {
	if value, exists := ctx.GetData(key); exists {
		if result, ok := value.(T); ok {
			return result
		}
	}
	var zero T
	return zero
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
