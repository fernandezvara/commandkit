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

// Set stores data in the context for middleware sharing
func (ctx *CommandContext) Set(key string, value any) {
	if ctx.data == nil {
		ctx.data = make(map[string]any)
	}
	ctx.data[key] = value
}

// Get retrieves data from the context
func (ctx *CommandContext) Get(key string) (any, bool) {
	if ctx.data == nil {
		return nil, false
	}
	value, exists := ctx.data[key]
	return value, exists
}

// GetString gets a string value from the context data
func (ctx *CommandContext) GetString(key string) string {
	if value, exists := ctx.Get(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt gets an int value from the context data
func (ctx *CommandContext) GetInt(key string) int {
	if value, exists := ctx.Get(key); exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool gets a bool value from the context data
func (ctx *CommandContext) GetBool(key string) bool {
	if value, exists := ctx.Get(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}
