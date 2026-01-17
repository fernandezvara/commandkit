// commandkit/files.go
package commandkit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// FileConfig represents configuration loaded from files
type FileConfig struct {
	data        map[string]any
	envPrefix   string
	environment string
}

// LoadFile loads configuration from a single file
func (c *Config) LoadFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config map[string]any
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".json":
		err = json.Unmarshal(data, &config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	case ".toml":
		err = toml.Unmarshal(data, &config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to parse %s file: %w", ext, err)
	}

	// Store file data for resolution
	if c.fileConfig == nil {
		c.fileConfig = &FileConfig{
			data: make(map[string]any),
		}
	}

	// Merge with existing file data
	c.mergeFileData(config)

	return nil
}

// LoadFiles loads configuration from multiple files (later files override earlier ones)
func (c *Config) LoadFiles(filenames ...string) error {
	for _, filename := range filenames {
		if err := c.LoadFile(filename); err != nil {
			return fmt.Errorf("error loading %s: %w", filename, err)
		}
	}
	return nil
}

// LoadFromEnv loads configuration file path from environment variable
func (c *Config) LoadFromEnv(envVar string) error {
	filename := os.Getenv(envVar)
	if filename == "" {
		return nil // No environment variable set
	}
	return c.LoadFile(filename)
}

// SetEnvironment sets the environment for environment-specific overrides
func (c *Config) SetEnvironment(env string) error {
	if c.fileConfig == nil {
		return fmt.Errorf("no configuration files loaded")
	}
	c.fileConfig.environment = env
	return nil
}

// SetEnvironmentFromEnv sets environment from environment variable
func (c *Config) SetEnvironmentFromEnv(envVar string) error {
	env := os.Getenv(envVar)
	if env == "" {
		return nil // No environment variable set
	}
	return c.SetEnvironment(env)
}

// WatchFile watches a configuration file for changes and reloads automatically
func (c *Config) WatchFile(filename string, callback func(error)) error {
	// For now, this is a placeholder. In a full implementation,
	// we would use fsnotify or similar to watch for file changes
	fmt.Printf("File watching not yet implemented for: %s\n", filename)
	return nil
}

// mergeFileData merges new config data with existing file data
func (c *Config) mergeFileData(newData map[string]any) {
	if c.fileConfig == nil {
		c.fileConfig = &FileConfig{
			data: make(map[string]any),
		}
	}

	// Simple merge - new data overrides old data
	for key, value := range newData {
		c.fileConfig.data[key] = value
	}
}

// getKeys returns all keys in a map (for debugging)
func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getFileValue gets a value from file configuration with environment support
func (c *Config) getFileValue(key string) (any, bool) {
	if c.fileConfig == nil {
		return nil, false
	}

	// Check for environment-specific override first
	if c.fileConfig.environment != "" {
		// Look for environments.{env}.{key} in nested structure
		if envs, exists := c.fileConfig.data["environments"]; exists {
			if envMap, ok := envs.(map[string]any); ok {
				if currentEnv, exists := envMap[c.fileConfig.environment]; exists {
					if currentEnvMap, ok := currentEnv.(map[string]any); ok {
						// Check exact key
						if value, exists := currentEnvMap[key]; exists {
							return value, true
						}
						// Check lowercase key
						if value, exists := currentEnvMap[strings.ToLower(key)]; exists {
							return value, true
						}
					}
				}
			}
		}
	}

	// Check for regular value (case-insensitive)
	if value, exists := c.fileConfig.data[key]; exists {
		return value, true
	}

	// Check lowercase version
	if value, exists := c.fileConfig.data[strings.ToLower(key)]; exists {
		return value, true
	}

	return nil, false
}

// Update resolveValue to include file configuration as highest priority
func (c *Config) resolveValueWithFiles(key string, def *Definition) (any, string, error) {
	var rawValue string
	var source string

	// Priority 1: Configuration files
	if fileValue, exists := c.getFileValue(key); exists {
		// Convert file value to string for parsing
		switch v := fileValue.(type) {
		case string:
			rawValue = v
		case bool:
			rawValue = fmt.Sprintf("%v", v)
		case int, int64, float64:
			rawValue = fmt.Sprintf("%v", v)
		case []any:
			// Handle arrays
			strs := make([]string, len(v))
			for i, item := range v {
				strs[i] = fmt.Sprintf("%v", item)
			}
			rawValue = strings.Join(strs, def.delimiter)
		default:
			return nil, "file", fmt.Errorf("unsupported file value type: %T", v)
		}
		source = "file"
	}

	// Priority 2: Environment variables
	if rawValue == "" && def.envVar != "" {
		if envVal := os.Getenv(def.envVar); envVal != "" {
			rawValue = envVal
			source = "env"
		}
	}

	// Priority 3: Command line flags
	if rawValue == "" && def.flag != "" {
		if flagVal, ok := c.flagValues[key]; ok && flagVal != nil && *flagVal != "" {
			rawValue = *flagVal
			source = "flag"
		}
	}

	// Priority 4: Default value
	if rawValue == "" && def.defaultValue != nil {
		source = "default"
		// Default is already the correct type, validate and return
		for _, v := range def.validations {
			if v.Name == "required" {
				continue // Skip required check for defaults
			}
			if err := v.Check(def.defaultValue); err != nil {
				return def.defaultValue, source, err
			}
		}
		return def.defaultValue, source, nil
	}

	// No value found
	if rawValue == "" {
		source = "none"
		if def.required {
			return nil, source, fmt.Errorf("required value not provided (set in file, %s or --%s)", def.envVar, def.flag)
		}
		return nil, source, nil
	}

	// Parse the raw string value into the expected type
	parsedValue, err := parseValue(rawValue, def.valueType, def.delimiter)
	if err != nil {
		return rawValue, source, err
	}

	// Run validations
	for _, validation := range def.validations {
		if err := validation.Check(parsedValue); err != nil {
			return parsedValue, source, err
		}
	}

	return parsedValue, source, nil
}
