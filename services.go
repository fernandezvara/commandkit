// commandkit/services.go
package commandkit

// CommandServices holds all command execution services
type CommandServices struct {
	Executor        CommandExecutor
	ConfigProcessor ConfigProcessor
	MiddlewareChain MiddlewareChain
	CommandRouter   CommandRouter
	FlagParser      FlagParser
}

// NewCommandServices creates a new CommandServices instance with all services initialized
func NewCommandServices() *CommandServices {
	return &CommandServices{
		Executor:        NewCommandExecutor(),
		ConfigProcessor: NewConfigProcessor(),
		MiddlewareChain: NewMiddlewareChain(),
		CommandRouter:   NewCommandRouter(),
		FlagParser:      NewFlagParser(),
	}
}
