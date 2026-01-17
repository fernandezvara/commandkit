package commandkit

import (
	"errors"
	"testing"
	"time"
)

func TestLoggingMiddleware(t *testing.T) {
	var loggedCtx *CommandContext
	var loggedDuration time.Duration

	middleware := LoggingMiddleware(func(ctx *CommandContext, duration time.Duration) {
		loggedCtx = ctx
		loggedDuration = duration
	})

	next := func(ctx *CommandContext) error {
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if loggedCtx == nil {
		t.Error("Expected context to be logged")
	}

	if loggedDuration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", loggedDuration)
	}

	if loggedCtx.Command != "test" {
		t.Errorf("Expected command 'test', got %s", loggedCtx.Command)
	}
}

func TestDefaultLoggingMiddleware(t *testing.T) {
	middleware := DefaultLoggingMiddleware()

	next := func(ctx *CommandContext) error {
		return nil
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAuthMiddleware(t *testing.T) {
	authCalled := false
	middleware := AuthMiddleware(func(ctx *CommandContext) error {
		authCalled = true
		return nil
	})

	next := func(ctx *CommandContext) error {
		return errors.New("command error")
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if !authCalled {
		t.Error("Expected auth function to be called")
	}

	if err == nil {
		t.Error("Expected error to be returned")
	}

	if err.Error() != "command error" {
		t.Errorf("Expected 'command error', got %v", err)
	}
}

func TestAuthMiddlewareFailure(t *testing.T) {
	middleware := AuthMiddleware(func(ctx *CommandContext) error {
		return errors.New("auth failed")
	})

	next := func(ctx *CommandContext) error {
		t.Error("Next function should not be called when auth fails")
		return nil
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if err == nil {
		t.Error("Expected authentication error")
	}

	if err.Error() != "authentication failed: auth failed" {
		t.Errorf("Expected 'authentication failed: auth failed', got %v", err)
	}
}

func TestTokenAuthMiddleware(t *testing.T) {
	cfg := New()
	cfg.Define("TOKEN").String().Default("valid-token")
	cfg.Process()

	middleware := TokenAuthMiddleware("TOKEN")

	next := func(ctx *CommandContext) error {
		// Check that token is stored in context
		if token, exists := ctx.Get("auth_token"); !exists || token.(string) != "valid-token" {
			t.Error("Expected auth token to be stored in context")
		}
		return nil
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	err := middleware(next)(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTokenAuthMiddlewareMissingToken(t *testing.T) {
	cfg := New()
	cfg.Define("TOKEN").String() // No default value
	cfg.Process()

	middleware := TokenAuthMiddleware("TOKEN")

	next := func(ctx *CommandContext) error {
		t.Error("Next function should not be called when token is missing")
		return nil
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	err := middleware(next)(ctx)

	if err == nil {
		t.Error("Expected error for missing token")
	}

	expected := "missing authentication token (config key: TOKEN)"
	if err.Error() != expected && err.Error() != "authentication failed: missing authentication token (config key: TOKEN)" {
		t.Errorf("Expected '%s' or 'authentication failed: %s', got %v", expected, expected, err)
	}
}

func TestErrorHandlingMiddleware(t *testing.T) {
	var handledErr error
	var handledCtx *CommandContext

	middleware := ErrorHandlingMiddleware(func(err error, ctx *CommandContext) {
		handledErr = err
		handledCtx = ctx
	})

	next := func(ctx *CommandContext) error {
		return errors.New("test error")
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if err == nil {
		t.Error("Expected error to be returned")
	}

	if handledErr == nil {
		t.Error("Expected error to be handled")
	}

	if handledCtx == nil {
		t.Error("Expected context to be passed to error handler")
	}

	// Check that error is stored in context
	if storedErr, exists := ctx.Get("error"); !exists || storedErr != handledErr {
		t.Error("Expected error to be stored in context")
	}
}

func TestTimingMiddleware(t *testing.T) {
	middleware := TimingMiddleware()

	next := func(ctx *CommandContext) error {
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that duration is stored in context
	if duration, exists := ctx.Get("duration"); !exists {
		t.Error("Expected duration to be stored in context")
	} else if duration.(time.Duration) < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", duration)
	}
}

func TestConditionalMiddleware(t *testing.T) {
	conditionMet := false
	middleware := ConditionalMiddleware(
		func(ctx *CommandContext) bool {
			return ctx.Command == "target"
		},
		func(next CommandFunc) CommandFunc {
			return func(ctx *CommandContext) error {
				conditionMet = true
				return next(ctx)
			}
		},
	)

	next := func(ctx *CommandContext) error {
		return nil
	}

	// Test condition met
	ctx1 := NewCommandContext([]string{}, New(), "target", "")
	err := middleware(next)(ctx1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !conditionMet {
		t.Error("Expected condition to be met and middleware to execute")
	}

	// Test condition not met
	conditionMet = false
	ctx2 := NewCommandContext([]string{}, New(), "other", "")
	err = middleware(next)(ctx2)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if conditionMet {
		t.Error("Expected condition not to be met and middleware to be skipped")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	middleware := RecoveryMiddleware()

	next := func(ctx *CommandContext) error {
		panic("test panic")
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	// Should not panic, but may return nil or error depending on implementation
	err := middleware(next)(ctx)

	// Check that panic is stored in context
	if panic, exists := ctx.Get("panic"); !exists || panic != "test panic" {
		t.Error("Expected panic to be stored in context")
	}

	// Recovery middleware should either return nil or an error, but not panic
	if err != nil {
		t.Logf("Recovery middleware returned error (this is acceptable): %v", err)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	middleware := RateLimitMiddleware(2, time.Minute)

	next := func(ctx *CommandContext) error {
		return nil
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	// First execution should succeed
	err := middleware(next)(ctx)
	if err != nil {
		t.Errorf("Unexpected error on first execution: %v", err)
	}

	// Second execution should succeed
	err = middleware(next)(ctx)
	if err != nil {
		t.Errorf("Unexpected error on second execution: %v", err)
	}

	// Third execution should fail
	err = middleware(next)(ctx)
	if err == nil {
		t.Error("Expected rate limit error on third execution")
	}

	expected := "rate limit exceeded: 2 executions allowed per 1m0s"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got %v", expected, err)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	var metricsCtx *CommandContext
	var metricsDuration time.Duration
	var metricsErr error

	middleware := MetricsMiddleware(func(ctx *CommandContext, duration time.Duration, err error) {
		metricsCtx = ctx
		metricsDuration = duration
		metricsErr = err
	})

	next := func(ctx *CommandContext) error {
		time.Sleep(10 * time.Millisecond) // Simulate work
		return errors.New("test error")
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if err == nil {
		t.Error("Expected error to be returned")
	}

	if metricsCtx == nil {
		t.Error("Expected metrics to be collected")
	}

	if metricsDuration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", metricsDuration)
	}

	if metricsErr == nil {
		t.Error("Expected error to be passed to metrics collector")
	}
}

func TestDefaultMetricsMiddleware(t *testing.T) {
	middleware := DefaultMetricsMiddleware()

	next := func(ctx *CommandContext) error {
		return nil
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")

	err := middleware(next)(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
