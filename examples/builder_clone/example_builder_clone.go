// Builder pattern example demonstrating Clone functionality and DRY patterns
package main

import (
	"fmt"
	"os"

	"github.com/fernandezvara/commandkit"
)

func main() {
	cfg := commandkit.New()

	// Example 1: Definition Builder Clone for variations
	fmt.Println("=== Example 1: Definition Builder Clone ===")

	// Create base port configuration
	basePortConfig := cfg.Define("PORT").
		Int64().
		Default(8080).
		Range(1, 65535).
		Description("Server port")

	// Clone and customize for different environments
	basePortConfig.Clone().
		Env("HTTP_PORT").
		Flag("http-port").
		Description("HTTP server port")

	basePortConfig.Clone().
		Env("HTTPS_PORT").
		Flag("https-port").
		Default(8443).
		Description("HTTPS server port")

	fmt.Printf("Port configurations created successfully\n")

	// Example 2: Command Builder Clone for DRY patterns
	fmt.Println("\n=== Example 2: Command Builder Clone (DRY Patterns) ===")

	// Create base configuration function
	baseServerConfig := func(cc *commandkit.CommandConfig) {
		cc.Define("PORT").
			Int64().
			Flag("port").
			Default(8080).
			Range(1, 65535).
			Description("Server port")

		cc.Define("HOST").
			String().
			Flag("host").
			Default("localhost").
			Description("Server host")

		cc.Define("VERBOSE").
			Bool().
			Flag("verbose").
			Default(false).
			Description("Enable verbose logging")
	}

	// Create base command template
	baseCommand := cfg.Command("base-template").
		ShortHelp("Base server command").
		Config(baseServerConfig)

	// Clone and customize for different server types
	baseCommand.Clone().
		ShortHelp("Start API server").
		Config(func(cc *commandkit.CommandConfig) {
			baseServerConfig(cc)
			cc.Define("API_KEY").
				String().
				Env("API_KEY").
				Required().
				Secret().
				Description("API authentication key")
		}).
		Func(apiServerCommand)

	baseCommand.Clone().
		ShortHelp("Start web server").
		Config(func(cc *commandkit.CommandConfig) {
			baseServerConfig(cc)
			cc.Define("STATIC_DIR").
				String().
				Flag("static-dir").
				Default("./static").
				Description("Static files directory")
		}).
		Func(webServerCommand)

	// Add commands to config (note: we need to add them with different names)
	cfg.Command("api-server").
		ShortHelp("Start API server").
		Config(func(cc *commandkit.CommandConfig) {
			baseServerConfig(cc)
			cc.Define("API_KEY").
				String().
				Env("API_KEY").
				Required().
				Secret().
				Description("API authentication key")
		}).
		Func(apiServerCommand)

	cfg.Command("web-server").
		ShortHelp("Start web server").
		Config(func(cc *commandkit.CommandConfig) {
			baseServerConfig(cc)
			cc.Define("STATIC_DIR").
				String().
				Flag("static-dir").
				Default("./static").
				Description("Static files directory")
		}).
		Func(webServerCommand)

	// Example 3: Complex command hierarchy with cloning
	fmt.Println("\n=== Example 3: Complex Command Hierarchy ===")

	// Create base admin command
	adminBase := cfg.Command("admin").
		ShortHelp("Administration commands").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("CONFIG_FILE").
				String().
				Flag("config").
				Default("/etc/app/config.yaml").
				Description("Configuration file path")
		})

	// Clone for different admin operations
	adminBase.Clone().
		ShortHelp("User administration").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("USER_FILE").
				String().
				Flag("users").
				Default("/etc/app/users.yaml").
				Description("Users configuration file")
		}).
		Func(adminUsersCommand)

	adminBase.Clone().
		ShortHelp("Database administration").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("DB_URL").
				String().
				Env("DATABASE_URL").
				Required().
				Secret().
				Description("Database connection URL")
			cc.Define("BACKUP_DIR").
				String().
				Flag("backup-dir").
				Default("/var/backups").
				Description("Backup directory")
		}).
		Func(adminDatabaseCommand)

	// Add the admin commands to config
	cfg.Command("admin-users").
		ShortHelp("User administration").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("CONFIG_FILE").
				String().
				Flag("config").
				Default("/etc/app/config.yaml").
				Description("Configuration file path")
			cc.Define("USER_FILE").
				String().
				Flag("users").
				Default("/etc/app/users.yaml").
				Description("Users configuration file")
		}).
		Func(adminUsersCommand)

	cfg.Command("admin-db").
		ShortHelp("Database administration").
		Config(func(cc *commandkit.CommandConfig) {
			cc.Define("CONFIG_FILE").
				String().
				Flag("config").
				Default("/etc/app/config.yaml").
				Description("Configuration file path")
			cc.Define("DB_URL").
				String().
				Env("DATABASE_URL").
				Required().
				Secret().
				Description("Database connection URL")
			cc.Define("BACKUP_DIR").
				String().
				Flag("backup-dir").
				Default("/var/backups").
				Description("Backup directory")
		}).
		Func(adminDatabaseCommand)

	fmt.Println("Commands created successfully!")
	fmt.Println("Commands: api-server, web-server, admin-users, admin-db")

	// Execute if arguments provided
	if len(os.Args) > 1 {
		if err := cfg.Execute(os.Args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func apiServerCommand(ctx *commandkit.CommandContext) error {
	fmt.Println("API Server starting...")

	portResult := commandkit.Get[int64](ctx, "PORT")
	if portResult.Error != nil {
		return fmt.Errorf("failed to get PORT: %w", portResult.Error)
	}
	port := commandkit.GetValue[int64](portResult)

	hostResult := commandkit.Get[string](ctx, "HOST")
	if hostResult.Error != nil {
		return fmt.Errorf("failed to get HOST: %w", hostResult.Error)
	}
	host := commandkit.GetValue[string](hostResult)

	apiKey := ctx.GlobalConfig.GetSecret("API_KEY")

	fmt.Printf("API Server starting on %s:%d\n", host, port)
	if apiKey.IsSet() {
		fmt.Printf("API Key configured (%d bytes)\n", apiKey.Size())
	}
	return nil
}

func webServerCommand(ctx *commandkit.CommandContext) error {
	fmt.Println("Web Server starting...")

	portResult := commandkit.Get[int64](ctx, "PORT")
	if portResult.Error != nil {
		return fmt.Errorf("failed to get PORT: %w", portResult.Error)
	}
	port := commandkit.GetValue[int64](portResult)

	hostResult := commandkit.Get[string](ctx, "HOST")
	if hostResult.Error != nil {
		return fmt.Errorf("failed to get HOST: %w", hostResult.Error)
	}
	host := commandkit.GetValue[string](hostResult)

	staticDirResult := commandkit.Get[string](ctx, "STATIC_DIR")
	if staticDirResult.Error != nil {
		return fmt.Errorf("failed to get STATIC_DIR: %w", staticDirResult.Error)
	}
	staticDir := commandkit.GetValue[string](staticDirResult)

	fmt.Printf("Web Server starting on %s:%d\n", host, port)
	fmt.Printf("Serving static files from: %s\n", staticDir)
	return nil
}

func adminUsersCommand(ctx *commandkit.CommandContext) error {
	fmt.Println("User Administration...")

	configFileResult := commandkit.Get[string](ctx, "CONFIG_FILE")
	if configFileResult.Error != nil {
		return fmt.Errorf("failed to get CONFIG_FILE: %w", configFileResult.Error)
	}
	configFile := commandkit.GetValue[string](configFileResult)

	userFileResult := commandkit.Get[string](ctx, "USER_FILE")
	if userFileResult.Error != nil {
		return fmt.Errorf("failed to get USER_FILE: %w", userFileResult.Error)
	}
	userFile := commandkit.GetValue[string](userFileResult)

	fmt.Printf("Config file: %s\n", configFile)
	fmt.Printf("User file: %s\n", userFile)
	fmt.Println("User administration completed")
	return nil
}

func adminDatabaseCommand(ctx *commandkit.CommandContext) error {
	fmt.Println("Database Administration...")

	configFileResult := commandkit.Get[string](ctx, "CONFIG_FILE")
	if configFileResult.Error != nil {
		return fmt.Errorf("failed to get CONFIG_FILE: %w", configFileResult.Error)
	}
	configFile := commandkit.GetValue[string](configFileResult)

	backupDirResult := commandkit.Get[string](ctx, "BACKUP_DIR")
	if backupDirResult.Error != nil {
		return fmt.Errorf("failed to get BACKUP_DIR: %w", backupDirResult.Error)
	}
	backupDir := commandkit.GetValue[string](backupDirResult)

	dbURL := ctx.GlobalConfig.GetSecret("DATABASE_URL")

	fmt.Printf("Config file: %s\n", configFile)
	fmt.Printf("Backup directory: %s\n", backupDir)
	if dbURL.IsSet() {
		fmt.Printf("Database connected (%d bytes)\n", dbURL.Size())
	}
	fmt.Println("Database administration completed")
	return nil
}
