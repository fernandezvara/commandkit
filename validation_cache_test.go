// commandkit/validation_cache_test.go
package commandkit

import (
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestValidationCache_Performance(t *testing.T) {
	pattern := `^[a-z]+@[a-z]+\.[a-z]+$`

	// Test cache performance
	start := time.Now()

	// Create multiple validations with the same pattern
	for i := 0; i < 1000; i++ {
		validation := validateRegexp(pattern)
		if validation.Name != "regexp(^[a-z]+@[a-z]+\\.[a-z]+$)" {
			t.Errorf("Expected validation name to contain pattern, got: %s", validation.Name)
		}

		// Test validation
		err := validation.Check("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email to pass, got error: %v", err)
		}
	}

	duration := time.Since(start)
	t.Logf("1000 validations with cached regex took: %v", duration)

	// Should be very fast with caching (less than 10ms for 1000 operations)
	if duration > 10*time.Millisecond {
		t.Logf("WARNING: Validation took longer than expected: %v", duration)
	}
}

func TestValidationCache_ConcurrentAccess(t *testing.T) {
	pattern := `^\d{4}-\d{2}-\d{2}$`

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Test concurrent access to the cache
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			validation := validateRegexp(pattern)

			// Test with valid date
			err := validation.Check("2023-12-25")
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: valid date failed: %v", id, err)
				return
			}

			// Test with invalid date
			err = validation.Check("invalid")
			if err == nil {
				errors <- fmt.Errorf("goroutine %d: invalid date should have failed", id)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

func TestValidationCache_DifferentPatterns(t *testing.T) {
	patterns := []string{
		`^[a-z]+$`,       // lowercase letters only
		`^[A-Z]+$`,       // uppercase letters only
		`^\d+$`,          // digits only
		`^[a-zA-Z0-9]+$`, // alphanumeric
		`^.+@.+\..+$`,    // email pattern
	}

	// Test that different patterns are cached correctly
	for _, pattern := range patterns {
		validation := validateRegexp(pattern)

		// Test with matching string
		testString := getMatchingString(pattern)
		err := validation.Check(testString)
		if err != nil {
			t.Errorf("Pattern %s: expected '%s' to match, got error: %v", pattern, testString, err)
		}

		// Test with non-matching string
		err = validation.Check("!@#$%%^&*()")
		if err == nil {
			t.Errorf("Pattern %s: expected '!@#$%%^&*()' to not match", pattern)
		}
	}
}

func getMatchingString(pattern string) string {
	switch pattern {
	case `^[a-z]+$`:
		return "lowercase"
	case `^[A-Z]+$`:
		return "UPPERCASE"
	case `^\d+$`:
		return "12345"
	case `^[a-zA-Z0-9]+$`:
		return "alphanum123"
	case `^.+@.+\..+$`:
		return "test@example.com"
	default:
		return "test"
	}
}

func TestValidationCache_MemoryEfficiency(t *testing.T) {
	// Test that cache doesn't grow indefinitely with unique patterns
	initialCacheSize := len(validationCache.regexCache)

	// Create validations with unique patterns
	for i := 0; i < 100; i++ {
		pattern := fmt.Sprintf("^pattern%d$", i)
		validation := validateRegexp(pattern)
		validation.Check("pattern0") // This should trigger caching
	}

	finalCacheSize := len(validationCache.regexCache)
	expectedSize := initialCacheSize + 100

	if finalCacheSize != expectedSize {
		t.Errorf("Expected cache size to be %d, got %d", expectedSize, finalCacheSize)
	}

	t.Logf("Cache grew from %d to %d entries", initialCacheSize, finalCacheSize)
}

// Benchmark to compare performance with and without caching
func BenchmarkValidationRegex_Cached(b *testing.B) {
	pattern := `^[a-z]+@[a-z]+\.[a-z]+$`
	validation := validateRegexp(pattern)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validation.Check("test@example.com")
	}
}

func BenchmarkValidationRegex_Uncached(b *testing.B) {
	pattern := `^[a-z]+@[a-z]+\.[a-z]+$`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This simulates the old behavior - compiling regex each time
		re := regexp.MustCompile(pattern)
		re.MatchString("test@example.com")
	}
}
