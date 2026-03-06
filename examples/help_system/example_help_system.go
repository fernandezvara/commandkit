// examples/help_system/example_help_system.go
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fernandezvara/commandkit"
)

func main() {
	// Create a new configuration with commands
	cfg := commandkit.New()

	// Define global configuration
	cfg.Define("ENV").String().Env("ENV").Default("dev").OneOf("dev", "staging", "prod").Description("Deployment environment")
	cfg.Define("DEBUG").Bool().Env("DEBUG").Default(false).Description("Enable debug mode")
	cfg.Define("TIMEOUT").Duration().Env("TIMEOUT").Default(30 * time.Second).Description("Request timeout")

	// Define commands with their own configurations
	cfg.Command("server").
		ShortHelp("Start the HTTP server").
		LongHelp("Start the HTTP server with configurable port and host. The server will listen on the specified address and serve the application.").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("HOST").String().Flag("host").Default("localhost").Description("Server host address")
			cc.Define("PORT").Int64().Flag("port").Default(8080).Range(1, 65535).Description("Server port")
			cc.Define("WORKERS").Int64().Flag("workers").Default(4).Range(1, 100).Description("Number of worker processes")
		}).
		Func(func(ctx *commandkit.CommandContext) error {
			fmt.Println("Server starting...")
			return nil
		})

	cfg.Command("deploy").
		ShortHelp("Deploy the application").
		LongHelp("Deploy the application to the specified environment with optional rollback capability.").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("TARGET").String().Flag("target").Required().Description("Deployment target environment")
			cc.Define("ROLLBACK").Bool().Flag("rollback").Default(false).Description("Enable automatic rollback on failure")
			cc.Define("API_KEY").String().Env("API_KEY").Required().Secret().Description("API authentication key")
		}).
		Func(func(ctx *commandkit.CommandContext) error {
			fmt.Println("Deploying...")
			return nil
		}).
		SubCommand("database").
			ShortHelp("Deploy database schema").
			Config(func(cc *commandkit.CommandConfig) {
				cc.Define("MIGRATE").Bool().Flag("migrate").Default(true).Description("Run database migrations")
				cc.Define("SEED").Bool().Flag("seed").Default(false).Description("Seed database with initial data")
			}).
			Func(func(ctx *commandkit.CommandContext) error {
				fmt.Println("Deploying database...")
				return nil
			})

	cfg.Command("admin").
		ShortHelp("Administrative operations").
		LongHelp("Perform administrative operations on the application, including user management and system maintenance.").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("FORCE").Bool().Flag("force").Default(false).Description("Force operation without confirmation")
		}).
		Func(func(ctx *commandkit.CommandContext) error {
			fmt.Println("Admin operation...")
			return nil
		}).
		SubCommand("users").
			ShortHelp("User management").
			Config(func(cc *commandkit.CommandConfig) {
				cc.Define("ACTION").String().Flag("action").Required().OneOf("list", "create", "delete", "update").Description("User action to perform")
				cc.Define("EMAIL").String().Flag("email").Description("User email for operations")
			}).
			Func(func(ctx *commandkit.CommandContext) error {
				fmt.Println("Managing users...")
				return nil
			})

	// Demonstrate the new help system
	demonstrateHelpSystem(cfg)
}

func demonstrateHelpSystem(cfg *commandkit.Config) {
	fmt.Println("=== CommandKit Help System Demonstration ===\n")

	// Create help integration
	integration := commandkit.NewHelpIntegration()

	// Add custom template functions
	integration.AddCustomFunction("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})

	integration.AddCustomFunction("banner", func(s string) string {
		lines := strings.Split(s, "\n")
		maxLen := 0
		for _, line := range lines {
			if len(line) > maxLen {
				maxLen = len(line)
			}
		}
		border := strings.Repeat("=", maxLen+4)
		result := border + "\n"
		for _, line := range lines {
			result += fmt.Sprintf("= %s =\n", line+strings.Repeat(" ", maxLen-len(line)))
		}
		result += border
		return result
	})

	// Set custom global template
	customGlobalTemplate := `{{banner "MyApp CLI Tool"}}

{{.Executable | upper}} - Command Line Interface

USAGE:
  {{.Executable}} <command> [options]

{{if .Commands}}AVAILABLE COMMANDS:
{{range .Commands}}{{if .Aliases}}  {{.Name | printf "%-12s"}} {{.Description}} ({{.Aliases | join ", "}})
{{else}}  {{.Name | printf "%-12s"}} {{.Description}}
{{end}}{{end}}{{end}}

EXAMPLES:
  {{.Executable}} server --host 0.0.0.0 --port 3000
  {{.Executable}} deploy --target production
  {{.Executable}} admin users --action list

For more information on a specific command: {{.Executable}} <command> --help`

	// Set custom command template
	customCommandTemplate := `{{banner (printf "%s COMMAND" .Command.Name | upper)}}

{{.Command.Name | upper}} - {{.Command.ShortHelp | title}}

{{if .Command.LongHelp}}{{.Command.LongHelp}}

{{end}}{{if .Usage}}USAGE:
  {{.Usage}}

{{end}}{{if .Flags}}OPTIONS:
{{range .Flags}}{{if .NoFlag}}  ENVIRONMENT: {{.EnvVar}}{{if .Required}} (REQUIRED){{end}}{{if .Default}} (default: {{.Default}}){{end}}
        {{.Description}}
{{else}}  --{{.Name}} {{.Type}}{{if .Required}} (REQUIRED){{end}}{{if .Default}} (default: {{.Default}}){{end}}{{if .EnvVar}} (env: {{.EnvVar}}){{end}}
        {{.Description}}
{{end}}{{end}}{{end}}{{if .Subcommands}}SUBCOMMANDS:
{{range .Subcommands}}  {{.Name | printf "%-12s"}} {{.Description}}{{if .Aliases}} ({{.Aliases | join ", "}}){{end}}
{{end}}{{end}}

{{if .Flags}}EXAMPLES:
  {{.Command.Name}} --{{(index .Flags 0).Name}} <value>
{{end}}For more help: {{.Command.Name}} <subcommand> --help`

	// Apply custom templates
	integration.SetCustomTemplate(commandkit.TemplateGlobal, customGlobalTemplate)
	integration.SetCustomTemplate(commandkit.TemplateCommand, customCommandTemplate)

	// Create string output for demonstration
	stringOutput := commandkit.NewStringHelpOutput()
	integration.SetOutput(stringOutput)

	// Get commands map for demonstration
	commands := getCommandsMap(cfg)

	// Demonstrate different help types
	fmt.Println("1. Global Help:")
	fmt.Println("==============")
	err := integration.ShowHelp([]string{"--help"}, commands)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println(stringOutput.Get())
	stringOutput.Reset()

	fmt.Println("\n2. Command Help (server):")
	fmt.Println("========================")
	err = integration.ShowHelp([]string{"server", "--help"}, commands)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println(stringOutput.Get())
	stringOutput.Reset()

	fmt.Println("\n3. Command Help (deploy):")
	fmt.Println("=======================")
	err = integration.ShowHelp([]string{"deploy", "--help"}, commands)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println(stringOutput.Get())
	stringOutput.Reset()

	fmt.Println("\n4. Subcommand Help (deploy database):")
	fmt.Println("=================================")
	err = integration.ShowHelp([]string{"deploy", "database", "--help"}, commands)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println(stringOutput.Get())
	stringOutput.Reset()

	fmt.Println("\n5. Help Generation (string output):")
	fmt.Println("===============================")
	text, err := integration.GenerateHelp([]string{"admin", "--help"}, commands)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(text)
	}

	// Demonstrate HelpConfig wrapper
	fmt.Println("\n6. HelpConfig Integration:")
	fmt.Println("======================")
	helpConfig := commandkit.NewHelpConfig(cfg)
	helpConfig.SetHelpOutput(commandkit.NewStringHelpOutput())

	fmt.Println("Global help via HelpConfig:")
	err = helpConfig.ShowGlobalHelp()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(helpConfig.GetHelpOutput().Get())
	}

	// Demonstrate HelpExecutor
	fmt.Println("\n7. HelpExecutor Integration:")
	fmt.Println("==========================")
	executor := commandkit.NewHelpExecutor()

	fmt.Println("Check and handle help:")
	handled, err := executor.CheckAndHandleHelp([]string{"--help"}, commands)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else if handled {
		fmt.Println("Help was displayed")
	} else {
		fmt.Println("No help requested")
	}

	fmt.Println("\n=== Help System Demonstration Complete ===")
}

// Helper function to extract commands map from Config
func getCommandsMap(cfg *commandkit.Config) map[string]*commandkit.Command {
	// This is a simplified approach - in a real implementation,
	// we would need to access the internal commands map
	// For now, we'll create a simple demonstration
	commands := make(map[string]*commandkit.Command)
	
	// Create mock commands for demonstration
	commands["server"] = &commandkit.Command{
		Name:        "server",
		ShortHelp:   "Start the HTTP server",
		LongHelp:    "Start the HTTP server with configurable port and host",
		Definitions: make(map[string]*commandkit.Definition),
		SubCommands: make(map[string]*commandkit.Command),
		Aliases:     []string{},
		Middleware:  []commandkit.CommandMiddleware{},
	}
	
	commands["deploy"] = &commandkit.Command{
		Name:        "deploy",
		ShortHelp:   "Deploy the application",
		LongHelp:    "Deploy the application to the specified environment",
		Definitions: make(map[string]*commandkit.Definition),
		SubCommands: map[string]*commandkit.Command{
			"database": {
				Name:        "database",
				ShortHelp:   "Deploy database schema",
				Definitions: make(map[string]*commandkit.Definition),
				SubCommands: make(map[string]*commandkit.Command),
				Aliases:     []string{},
				Middleware:  []commandkit.CommandMiddleware{},
			},
		},
		Aliases:    []string{},
		Middleware: []commandkit.CommandMiddleware{},
	}
	
	commands["admin"] = &commandkit.Command{
		Name:        "admin",
		ShortHelp:   "Administrative operations",
		LongHelp:    "Perform administrative operations on the application",
		Definitions: make(map[string]*commandkit.Definition),
		SubCommands: map[string]*commandkit.Command{
			"users": {
				Name:        "users",
				ShortHelp:   "User management",
				Definitions: make(map[string]*commandkit.Definition),
				SubCommands: make(map[string]*commandkit.Command),
				Aliases:     []string{},
				Middleware:  []commandkit.CommandMiddleware{},
			},
		},
		Aliases:    []string{},
		Middleware: []commandkit.CommandMiddleware{},
	}
	
	return commands
}
