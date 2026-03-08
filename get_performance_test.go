package commandkit

import (
	"testing"
)

// TestGetPerformance tests the performance optimizations in Get[T]
func TestGetPerformance(t *testing.T) {
	cfg := New()
	cfg.Define("PORT").Int64().Default(int64(8080))
	cfg.Define("HOST").String().Default("localhost")
	cfg.Define("DEBUG").Bool().Default(false)
	cfg.Define("RATE").Float64().Default(0.5)

	if err := cfg.Execute([]string{"test"}); err != nil {
		t.Fatalf("Failed to process config: %v", err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test multiple Get calls to verify type caching
	for i := 0; i < 100; i++ {
		// Test int64
		port, err := Get[int64](ctx, "PORT")
		if err != nil {
			t.Errorf("Failed to get PORT: %v", err)
		}
		if port != 8080 {
			t.Errorf("Expected PORT=8080, got %d", port)
		}

		// Test string
		host, err := Get[string](ctx, "HOST")
		if err != nil {
			t.Errorf("Failed to get HOST: %v", err)
		}
		if host != "localhost" {
			t.Errorf("Expected HOST=localhost, got %s", host)
		}

		// Test bool
		debug, err := Get[bool](ctx, "DEBUG")
		if err != nil {
			t.Errorf("Failed to get DEBUG: %v", err)
		}
		if debug != false {
			t.Errorf("Expected DEBUG=false, got %v", debug)
		}

		// Test float64
		rate, err := Get[float64](ctx, "RATE")
		if err != nil {
			t.Errorf("Failed to get RATE: %v", err)
		}
		if rate != 0.5 {
			t.Errorf("Expected RATE=0.5, got %f", rate)
		}
	}
}

// BenchmarkGetTypeDescription benchmarks the type description caching
func BenchmarkGetTypeDescription(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		typeDescription("test")
		typeDescription(int64(0))
		typeDescription(int(0))
		typeDescription(true)
		typeDescription(float64(0))
		typeDescription([]string{"a", "b"})
		typeDescription([]int64{1, 2})
		typeDescription([]int{1, 2})
	}
}

// BenchmarkGetInt64 benchmarks Get[int64] performance
func BenchmarkGetInt64(b *testing.B) {
	cfg := New()
	cfg.Define("PORT").Int64().Default(int64(8080))

	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Get[int64](ctx, "PORT")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetString benchmarks Get[string] performance
func BenchmarkGetString(b *testing.B) {
	cfg := New()
	cfg.Define("HOST").String().Default("localhost")

	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Get[string](ctx, "HOST")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetBool benchmarks Get[bool] performance
func BenchmarkGetBool(b *testing.B) {
	cfg := New()
	cfg.Define("DEBUG").Bool().Default(false)

	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Get[bool](ctx, "DEBUG")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetInt benchmarks Get[int] performance
func BenchmarkGetInt(b *testing.B) {
	cfg := New()
	cfg.Define("PORT").Int().Default(8080)

	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Get[int](ctx, "PORT")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetStringSlice benchmarks Get[[]string] performance
func BenchmarkGetStringSlice(b *testing.B) {
	cfg := New()
	cfg.Define("TAGS").StringSlice().Default([]string{"tag1", "tag2"})

	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Get[[]string](ctx, "TAGS")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetInt64Slice benchmarks Get[[]int64] performance
func BenchmarkGetInt64Slice(b *testing.B) {
	cfg := New()
	cfg.Define("NUMBERS").Int64Slice().Default([]int64{1, 2, 3})

	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Get[[]int64](ctx, "NUMBERS")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetIntSlice benchmarks Get[[]int] performance
func BenchmarkGetIntSlice(b *testing.B) {
	cfg := New()
	cfg.Define("PORTS").IntSlice().Default([]int{8080, 8081})

	if err := cfg.Execute([]string{"test"}); err != nil {
		b.Fatal(err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Get[[]int](ctx, "PORTS")
		if err != nil {
			b.Fatal(err)
		}
	}
}
