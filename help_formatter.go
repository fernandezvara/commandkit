// commandkit/help_formatter.go
package commandkit

import (
	"fmt"
	"sort"
	"strings"
)

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

// Enhanced formatting functions for templates
func (hf *TemplateHelpFormatter) formatCommandList(commands []CommandSummary) string {
	var sb strings.Builder
	
	for _, cmd := range commands {
		aliases := ""
		if len(cmd.Aliases) > 0 {
			aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(cmd.Aliases, ", "))
		}
		fmt.Fprintf(&sb, "  %-12s %s%s\n", cmd.Name, cmd.Description, aliases)
	}
	
	return sb.String()
}

func (hf *TemplateHelpFormatter) formatFlagList(flags []FlagInfo) string {
	var sb strings.Builder
	
	for _, flag := range flags {
		if flag.NoFlag {
			// Environment-only configuration
			fmt.Fprintf(&sb, "  (no flag) %s", flag.Type)
			
			var indicators []string
			if flag.EnvVar != "" {
				indicators = append(indicators, fmt.Sprintf("env: %s", flag.EnvVar))
			}
			if flag.Required {
				indicators = append(indicators, "required")
			}
			if flag.Default != nil {
				if flag.Secret {
					indicators = append(indicators, "default: [hidden]")
				} else {
					indicators = append(indicators, fmt.Sprintf("default: %v", flag.Default))
				}
			}
			if flag.Secret {
				indicators = append(indicators, "secret")
			}
			if len(indicators) > 0 {
				fmt.Fprintf(&sb, " (%s)", strings.Join(indicators, ", "))
			}
			
			fmt.Fprintf(&sb, "\n        %s\n", flag.Description)
		} else {
			// Regular flag
			flagName := flag.Name
			if flagName == "" {
				flagName = strings.ToLower(strings.ReplaceAll(flag.Name, "_", "-"))
			}
			
			fmt.Fprintf(&sb, "  --%s %s", flagName, flag.Type)
			
			var indicators []string
			if flag.Required {
				indicators = append(indicators, "required")
			}
			if flag.Default != nil {
				if flag.Secret {
					indicators = append(indicators, "default: [hidden]")
				} else {
					if flag.Type == "string" {
						indicators = append(indicators, fmt.Sprintf("default: '%v'", flag.Default))
					} else {
						indicators = append(indicators, fmt.Sprintf("default: %v", flag.Default))
					}
				}
			}
			if flag.EnvVar != "" {
				indicators = append(indicators, fmt.Sprintf("env: %s", flag.EnvVar))
			}
			if len(flag.Validations) > 0 {
				indicators = append(indicators, strings.Join(flag.Validations, ", "))
			}
			if flag.Secret {
				indicators = append(indicators, "secret")
			}
			if len(indicators) > 0 {
				fmt.Fprintf(&sb, " (%s)", strings.Join(indicators, ", "))
			}
			
			fmt.Fprintf(&sb, "\n        %s\n", flag.Description)
		}
	}
	
	return sb.String()
}

func (hf *TemplateHelpFormatter) formatSubcommandList(subcommands []SubcommandInfo) string {
	var sb strings.Builder
	
	for _, subcmd := range subcommands {
		aliases := ""
		if len(subcmd.Aliases) > 0 {
			aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(subcmd.Aliases, ", "))
		}
		fmt.Fprintf(&sb, "  %-12s %s%s\n", subcmd.Name, subcmd.Description, aliases)
	}
	
	return sb.String()
}

// Helper functions for template data preparation
func (hf *TemplateHelpFormatter) prepareCommandData(help *CommandHelp) map[string]interface{} {
	return map[string]interface{}{
		"Command":     help.Command,
		"Usage":       help.Usage,
		"Description": help.Description,
		"Flags":       help.Flags,
		"Subcommands": help.Subcommands,
		"HasFlags":    len(help.Flags) > 0,
		"HasSubcommands": len(help.Subcommands) > 0,
		"FlagCount":   len(help.Flags),
		"SubcommandCount": len(help.Subcommands),
	}
}

func (hf *TemplateHelpFormatter) prepareGlobalData(help *GlobalHelp) map[string]interface{} {
	return map[string]interface{}{
		"Executable": help.Executable,
		"Commands":   help.Commands,
		"CommandCount": len(help.Commands),
		"HasCommands": len(help.Commands) > 0,
	}
}

func (hf *TemplateHelpFormatter) prepareSubcommandData(help *SubcommandHelp) map[string]interface{} {
	return map[string]interface{}{
		"Parent":      help.Parent,
		"Subcommands": help.Subcommands,
		"SubcommandCount": len(help.Subcommands),
		"HasSubcommands": len(help.Subcommands) > 0,
	}
}

func (hf *TemplateHelpFormatter) prepareFlagData(help *FlagHelp) map[string]interface{} {
	return map[string]interface{}{
		"Command":    help.Command,
		"Flags":      help.Flags,
		"FlagCount":  len(help.Flags),
		"HasFlags":   len(help.Flags) > 0,
	}
}

// Sort helpers for templates
func (hf *TemplateHelpFormatter) sortCommands(commands []CommandSummary) []CommandSummary {
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	return commands
}

func (hf *TemplateHelpFormatter) sortFlags(flags []FlagInfo) []FlagInfo {
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Name < flags[j].Name
	})
	return flags
}

func (hf *TemplateHelpFormatter) sortSubcommands(subcommands []SubcommandInfo) []SubcommandInfo {
	sort.Slice(subcommands, func(i, j int) bool {
		return subcommands[i].Name < subcommands[j].Name
	})
	return subcommands
}
