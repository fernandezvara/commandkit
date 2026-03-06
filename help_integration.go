// commandkit/help_integration.go
package commandkit

// HelpIntegration provides integration between the new help system and existing CommandKit
type HelpIntegration struct {
	helpService HelpService
}

// NewHelpIntegration creates a new help integration
func NewHelpIntegration() *HelpIntegration {
	return &HelpIntegration{
		helpService: NewHelpService(),
	}
}

// ShowHelp replaces the old help system with the new template-based help
func (hi *HelpIntegration) ShowHelp(args []string, commands map[string]*Command) error {
	return hi.helpService.ShowHelp(args, commands)
}

// GenerateHelp generates help text using the new system
func (hi *HelpIntegration) GenerateHelp(args []string, commands map[string]*Command) (string, error) {
	return hi.helpService.GenerateHelp(args, commands)
}

// GetHelpService returns the underlying help service for advanced usage
func (hi *HelpIntegration) GetHelpService() HelpService {
	return hi.helpService
}

// SetCustomTemplate allows setting custom templates for branding/formatting
func (hi *HelpIntegration) SetCustomTemplate(templateType TemplateType, template string) {
	formatter := hi.helpService.GetFormatter()
	if templateFormatter, ok := formatter.(*TemplateHelpFormatter); ok {
		templateFormatter.SetTemplate(templateType, template)
	}
}

// AddCustomFunction adds a custom function to the template renderer
func (hi *HelpIntegration) AddCustomFunction(name string, fn interface{}) {
	formatter := hi.helpService.GetFormatter()
	if templateFormatter, ok := formatter.(*TemplateHelpFormatter); ok {
		renderer := templateFormatter.GetRenderer()
		renderer.AddFunction(name, fn)
	}
}

// SetOutput sets the output destination for help
func (hi *HelpIntegration) SetOutput(output HelpOutput) {
	hi.helpService.SetOutput(output)
}

// GetOutput returns the current output destination
func (hi *HelpIntegration) GetOutput() HelpOutput {
	return hi.helpService.GetOutput()
}

// HelpConfig provides help functionality for global Config
type HelpConfig struct {
	*Config
	helpIntegration *HelpIntegration
}

// NewHelpConfig wraps a Config with help functionality
func NewHelpConfig(config *Config) *HelpConfig {
	return &HelpConfig{
		Config:          config,
		helpIntegration: NewHelpIntegration(),
	}
}

// ShowGlobalHelp shows help for all commands
func (hc *HelpConfig) ShowGlobalHelp() error {
	return hc.helpIntegration.ShowHelp([]string{"--help"}, hc.commands)
}

// ShowCommandHelp shows help for a specific command
func (hc *HelpConfig) ShowCommandHelp(commandName string) error {
	return hc.helpIntegration.ShowHelp([]string{commandName, "--help"}, hc.commands)
}

// GenerateGlobalHelpText generates global help text
func (hc *HelpConfig) GenerateGlobalHelpText() (string, error) {
	return hc.helpIntegration.GenerateHelp([]string{"--help"}, hc.commands)
}

// GenerateCommandHelpText generates command help text
func (hc *HelpConfig) GenerateCommandHelpText(commandName string) (string, error) {
	return hc.helpIntegration.GenerateHelp([]string{commandName, "--help"}, hc.commands)
}

// SetHelpOutput sets the output destination for help
func (hc *HelpConfig) SetHelpOutput(output HelpOutput) {
	hc.helpIntegration.SetOutput(output)
}

// GetHelpOutput returns the current help output destination
func (hc *HelpConfig) GetHelpOutput() HelpOutput {
	return hc.helpIntegration.GetOutput()
}

// HelpExecutor provides help functionality for command execution
type HelpExecutor struct {
	helpIntegration *HelpIntegration
}

// NewHelpExecutor creates a new help executor
func NewHelpExecutor() *HelpExecutor {
	return &HelpExecutor{
		helpIntegration: NewHelpIntegration(),
	}
}

// ExecuteHelp handles help requests during command execution
func (he *HelpExecutor) ExecuteHelp(args []string, commands map[string]*Command) error {
	return he.helpIntegration.ShowHelp(args, commands)
}

// CheckAndHandleHelp checks if help is requested and handles it
func (he *HelpExecutor) CheckAndHandleHelp(args []string, commands map[string]*Command) (bool, error) {
	if !he.helpIntegration.GetHelpService().IsHelpRequested(args) {
		return false, nil
	}

	err := he.helpIntegration.ShowHelp(args, commands)
	return true, err
}

// GetHelpExecutor returns the help executor
func (hi *HelpIntegration) GetHelpExecutor() *HelpExecutor {
	return &HelpExecutor{
		helpIntegration: hi,
	}
}

// GetFormatter returns the help formatter for advanced customization
func (hi *HelpIntegration) GetFormatter() HelpFormatter {
	return hi.helpService.GetFormatter()
}

// GetFactory returns the help factory for advanced usage
func (hi *HelpIntegration) GetFactory() HelpFactory {
	return hi.helpService.GetFactory()
}
