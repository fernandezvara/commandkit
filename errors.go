// commandkit/errors.go
package commandkit

import (
	"fmt"
	"strings"
)

// ConfigError represents a single configuration error
type ConfigError struct {
	Key     string
	Source  string // "env", "flag", "default", or "none"
	Value   string // Masked if secret
	Message string
}

func (e *ConfigError) Error() string {
	if e.Source == "none" {
		return fmt.Sprintf("%s: %s", e.Key, e.Message)
	}
	if e.Value != "" {
		return fmt.Sprintf("%s (%s=%s): %s", e.Key, e.Source, e.Value, e.Message)
	}
	return fmt.Sprintf("%s (%s): %s", e.Key, e.Source, e.Message)
}

// formatErrors creates a nicely formatted error output
func formatErrors(errs []ConfigError) string {
	if len(errs) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                    CONFIGURATION ERRORS                          ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")

	for i, err := range errs {
		// Key line
		sb.WriteString(fmt.Sprintf("║  %-64s║\n", fmt.Sprintf("❌ %s", err.Key)))

		// Source and value
		if err.Source != "none" {
			sourceInfo := fmt.Sprintf("   Source: %s", err.Source)
			if err.Value != "" {
				sourceInfo += fmt.Sprintf(" = %s", err.Value)
			}
			sb.WriteString(fmt.Sprintf("║  %-64s║\n", sourceInfo))
		}

		// Error message
		sb.WriteString(fmt.Sprintf("║  %-64s║\n", fmt.Sprintf("   Error: %s", err.Message)))

		// Separator between errors
		if i < len(errs)-1 {
			sb.WriteString("║  ────────────────────────────────────────────────────────────    ║\n")
		}
	}

	sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║  %-64s║\n", fmt.Sprintf("Total: %d error(s)", len(errs))))
	sb.WriteString("╚══════════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// maskSecret masks a secret value for display
func maskSecret(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
