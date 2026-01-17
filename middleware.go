// commandkit/middleware.go
package commandkit

import (
	"fmt"
	"log"
	"time"
)

// LoggingMiddleware creates middleware that logs command execution with timing
// The logger function receives the command context and execution duration
func LoggingMiddleware(logger func(*CommandContext, time.Duration)) CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			start := time.Now()

			// Log command start
			log.Printf("🚀 Starting command: %s", ctx.Command)
			if ctx.SubCommand != "" {
				log.Printf("📂 Subcommand: %s", ctx.SubCommand)
			}

			// Execute next in chain
			err := next(ctx)

			// Log completion with timing
			duration := time.Since(start)
			logger(ctx, duration)

			return err
		}
	}
}

// DefaultLoggingMiddleware creates a standard logging middleware with sensible defaults
func DefaultLoggingMiddleware() CommandMiddleware {
	return LoggingMiddleware(func(ctx *CommandContext, duration time.Duration) {
		status := "✅ SUCCESS"
		if _, hasError := ctx.Get("error"); hasError {
			status = "❌ FAILED"
		}

		log.Printf("%s Command %s completed in %v", status, ctx.Command, duration)
	})
}

// AuthMiddleware creates middleware that validates authentication before command execution
// The auth function should return nil if authentication succeeds, or an error if it fails
func AuthMiddleware(authFunc func(*CommandContext) error) CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			// Check authentication before executing command
			if err := authFunc(ctx); err != nil {
				log.Printf("🔒 Authentication failed for command %s: %v", ctx.Command, err)
				return fmt.Errorf("authentication failed: %w", err)
			}

			log.Printf("🔓 Authentication successful for command %s", ctx.Command)

			// Auth passed, execute command
			return next(ctx)
		}
	}
}

// TokenAuthMiddleware creates authentication middleware that validates a token from config
// It looks for the token in the config using the provided key
func TokenAuthMiddleware(tokenKey string) CommandMiddleware {
	return AuthMiddleware(func(ctx *CommandContext) error {
		var token string

		// Check if token exists and get it appropriately
		if !ctx.Config.Has(tokenKey) {
			return fmt.Errorf("missing authentication token (config key: %s)", tokenKey)
		}

		// Check if this is defined as a secret
		if ctx.Config.IsSecret(tokenKey) {
			secret := ctx.Config.GetSecret(tokenKey)
			if !secret.IsSet() {
				return fmt.Errorf("missing authentication token (config key: %s)", tokenKey)
			}
			token = secret.String()
		} else {
			token = ctx.Config.GetString(tokenKey)
		}

		if token == "" {
			return fmt.Errorf("missing authentication token (config key: %s)", tokenKey)
		}

		// Add token to context for potential use by other middleware/commands
		ctx.Set("auth_token", token)

		return nil
	})
}

// ErrorHandlingMiddleware creates middleware that handles errors from command execution
// The errorHandler function receives the error and command context for logging/monitoring
func ErrorHandlingMiddleware(errorHandler func(error, *CommandContext)) CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			err := next(ctx)

			if err != nil {
				// Store error in context for other middleware
				ctx.Set("error", err)

				// Handle the error (logging, monitoring, etc.)
				errorHandler(err, ctx)
			}

			return err
		}
	}
}

// DefaultErrorHandlingMiddleware creates standard error handling with logging
func DefaultErrorHandlingMiddleware() CommandMiddleware {
	return ErrorHandlingMiddleware(func(err error, ctx *CommandContext) {
		log.Printf("💥 Error in command %s: %v", ctx.Command, err)

		// You could add monitoring integration here:
		// monitor.Error("command_failed", map[string]any{
		//     "command": ctx.Command,
		//     "error": err.Error(),
		// })
	})
}

// TimingMiddleware creates middleware that measures and stores execution timing
// The timing is stored in ctx.Values["duration"] for other middleware to use
func TimingMiddleware() CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			start := time.Now()
			err := next(ctx)
			duration := time.Since(start)

			// Store timing in context for other middleware
			ctx.Set("duration", duration)

			log.Printf("⏱️ Command %s took %v", ctx.Command, duration)

			return err
		}
	}
}

// ConditionalMiddleware creates middleware that only applies when the condition is true
// This allows for sophisticated conditional middleware application
func ConditionalMiddleware(condition func(*CommandContext) bool, middleware CommandMiddleware) CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			if condition(ctx) {
				log.Printf("🔧 Applying conditional middleware for command %s", ctx.Command)
				return middleware(next)(ctx)
			}
			log.Printf("⏭️ Skipping conditional middleware for command %s", ctx.Command)
			return next(ctx)
		}
	}
}

// AdminOnlyMiddleware creates middleware that only allows admin commands
// It checks for an admin token in the configuration
func AdminOnlyMiddleware(adminTokenKey string) CommandMiddleware {
	return ConditionalMiddleware(
		func(ctx *CommandContext) bool {
			// Only apply to commands starting with "admin-"
			return len(ctx.Command) > 6 && ctx.Command[:6] == "admin-"
		},
		AuthMiddleware(func(ctx *CommandContext) error {
			token := ctx.Config.GetString(adminTokenKey)
			if token == "" {
				return fmt.Errorf("admin commands require authentication token (config key: %s)", adminTokenKey)
			}

			// Additional admin validation could go here
			if token != "admin-secret" && token != "admin-token" {
				return fmt.Errorf("invalid admin token")
			}

			return nil
		}),
	)
}

// RecoveryMiddleware creates middleware that recovers from panics
// This prevents the entire application from crashing due to panics in commands
func RecoveryMiddleware() CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("🚨 Panic recovered in command %s: %v", ctx.Command, r)

					// Store panic in context for error handling middleware
					ctx.Set("panic", r)
				}
			}()

			return next(ctx)
		}
	}
}

// RateLimitMiddleware creates middleware that implements basic rate limiting
// It tracks command execution count in the context
func RateLimitMiddleware(maxExecutions int, window time.Duration) CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			// Initialize rate limit tracking in context

			// Get current execution count
			var count int
			if c, exists := ctx.Get("execution_count"); exists {
				count = c.(int)
			}

			count++
			ctx.Set("execution_count", count)

			if count > maxExecutions {
				return fmt.Errorf("rate limit exceeded: %d executions allowed per %v", maxExecutions, window)
			}

			log.Printf("📊 Command %s execution count: %d/%d", ctx.Command, count, maxExecutions)

			return next(ctx)
		}
	}
}

// MetricsMiddleware creates middleware for collecting command metrics
// This is useful for monitoring and analytics
func MetricsMiddleware(metricsCollector func(*CommandContext, time.Duration, error)) CommandMiddleware {
	return func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			start := time.Now()
			err := next(ctx)
			duration := time.Since(start)

			// Collect metrics
			metricsCollector(ctx, duration, err)

			return err
		}
	}
}

// DefaultMetricsMiddleware creates standard metrics collection
func DefaultMetricsMiddleware() CommandMiddleware {
	return MetricsMiddleware(func(ctx *CommandContext, duration time.Duration, err error) {
		status := "success"
		if err != nil {
			status = "error"
		}

		log.Printf("📈 Metrics: command=%s duration=%v status=%s", ctx.Command, duration, status)

		// In a real application, you'd send this to a metrics system:
		// metrics.Counter("command_executions", map[string]string{
		//     "command": ctx.Command,
		//     "status":  status,
		// }).Inc()
		//
		// metrics.Histogram("command_duration", map[string]string{
		//     "command": ctx.Command,
		// }).Observe(duration.Seconds())
	})
}
