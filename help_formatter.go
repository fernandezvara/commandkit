// commandkit/help_formatter.go
package commandkit

// HelpFormatter formats help objects into strings using templates
type HelpFormatter interface {
	FormatGlobalHelp(help *GlobalHelp) (string, error)
	FormatCommandHelp(help *CommandHelp) (string, error)
	FormatSubcommandHelp(help *SubcommandHelp) (string, error)
	FormatFlagHelp(help *FlagHelp) (string, error)
	SetTemplate(templateType TemplateType, template string)
	GetTemplate(templateType TemplateType) string
	SetRenderer(renderer TemplateRenderer)
	GetRenderer() TemplateRenderer
}

// TemplateHelpFormatter implements HelpFormatter with template support
type TemplateHelpFormatter struct {
	templates map[TemplateType]string
	renderer  TemplateRenderer
}

// NewTemplateHelpFormatter creates a new template-based help formatter
func NewTemplateHelpFormatter() HelpFormatter {
	formatter := &TemplateHelpFormatter{
		templates: make(map[TemplateType]string),
		renderer:  NewGoTemplateRenderer(),
	}

	// Set default templates
	formatter.setDefaultTemplates()

	return formatter
}

// FormatGlobalHelp formats global help using templates
func (hf *TemplateHelpFormatter) FormatGlobalHelp(help *GlobalHelp) (string, error) {
	templateStr := help.Template
	if templateStr == "" {
		templateStr = hf.templates[TemplateGlobal]
	}

	data := map[string]interface{}{
		"Executable": help.Executable,
		"Commands":   help.Commands,
	}

	return hf.renderer.Render(templateStr, data)
}

// FormatCommandHelp formats command help using templates
func (hf *TemplateHelpFormatter) FormatCommandHelp(help *CommandHelp) (string, error) {
	templateStr := help.Template
	if templateStr == "" {
		templateStr = hf.templates[TemplateCommand]
	}

	data := map[string]interface{}{
		"Command":     help.Command,
		"Usage":       help.Usage,
		"Description": help.Description,
		"Flags":       help.Flags,
		"Subcommands": help.Subcommands,
	}

	return hf.renderer.Render(templateStr, data)
}

// FormatSubcommandHelp formats subcommand help using templates
func (hf *TemplateHelpFormatter) FormatSubcommandHelp(help *SubcommandHelp) (string, error) {
	templateStr := help.Template
	if templateStr == "" {
		templateStr = hf.templates[TemplateSubcommand]
	}

	data := map[string]interface{}{
		"Parent":      help.Parent,
		"Subcommands": help.Subcommands,
	}

	return hf.renderer.Render(templateStr, data)
}

// FormatFlagHelp formats flag help using templates
func (hf *TemplateHelpFormatter) FormatFlagHelp(help *FlagHelp) (string, error) {
	templateStr := help.Template
	if templateStr == "" {
		templateStr = hf.templates[TemplateFlag]
	}

	data := map[string]interface{}{
		"Command": help.Command,
		"Flags":   help.Flags,
	}

	return hf.renderer.Render(templateStr, data)
}

// SetTemplate sets a custom template for a specific type
func (hf *TemplateHelpFormatter) SetTemplate(templateType TemplateType, template string) {
	hf.templates[templateType] = template
}

// GetTemplate gets the current template for a specific type
func (hf *TemplateHelpFormatter) GetTemplate(templateType TemplateType) string {
	return hf.templates[templateType]
}

// SetRenderer sets the template renderer
func (hf *TemplateHelpFormatter) SetRenderer(renderer TemplateRenderer) {
	hf.renderer = renderer
}

// GetRenderer returns the current template renderer
func (hf *TemplateHelpFormatter) GetRenderer() TemplateRenderer {
	return hf.renderer
}

// setDefaultTemplates sets the default templates
func (hf *TemplateHelpFormatter) setDefaultTemplates() {
	hf.templates[TemplateGlobal] = DefaultGlobalTemplate
	hf.templates[TemplateCommand] = DefaultCommandTemplate
	hf.templates[TemplateSubcommand] = DefaultSubcommandTemplate
	hf.templates[TemplateFlag] = DefaultFlagTemplate
}
