// CommandKit Type Conversion TDD Test Suite
// Tests all type conversions across all sources (DEFAULT, ENV, FLAG, FILE)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fernandezvara/commandkit"
	"github.com/google/uuid"
)

// TestResult holds one test outcome
type TestResult struct {
	Source       string
	Key          string
	ExpectedType string
	ReturnedType string
	Value        string
	Success      bool
	Error        string
}

var allResults []TestResult

func addResult(source, key, expectedType, returnedType, value string, success bool, errStr string) {
	allResults = append(allResults, TestResult{
		Source:       source,
		Key:          key,
		ExpectedType: expectedType,
		ReturnedType: returnedType,
		Value:        value,
		Success:      success,
		Error:        errStr,
	})
}

// testGet is a generic helper that tests Get[T] and records the result
func testGet[T any](source, key, expectedType string, ctx *commandkit.CommandContext) {
	val, err := commandkit.Get[T](ctx, key)
	if err != nil {
		addResult(source, key, expectedType, "error", "", false, err.Error())
		return
	}
	valStr := fmt.Sprintf("%v", val)
	retType := fmt.Sprintf("%T", val)
	addResult(source, key, expectedType, retType, valStr, true, "")
}

// runAllTypeTests runs Get[T] for every type using the given context
func runAllTypeTests(source string, ctx *commandkit.CommandContext) {
	// Basic types
	testGet[string](source, "TEST_STRING", "string", ctx)
	testGet[bool](source, "TEST_BOOL", "bool", ctx)

	// Integer types
	testGet[int64](source, "TEST_INT64", "int64", ctx)
	testGet[int](source, "TEST_INT", "int", ctx)
	testGet[uint](source, "TEST_UINT", "uint", ctx)
	testGet[uint8](source, "TEST_UINT8", "uint8", ctx)
	testGet[uint16](source, "TEST_UINT16", "uint16", ctx)
	testGet[uint32](source, "TEST_UINT32", "uint32", ctx)
	testGet[uint64](source, "TEST_UINT64", "uint64", ctx)

	// Float types
	testGet[float64](source, "TEST_FLOAT64", "float64", ctx)
	testGet[float32](source, "TEST_FLOAT32", "float32", ctx)

	// Time types
	testGet[time.Duration](source, "TEST_DURATION", "time.Duration", ctx)
	testGet[time.Time](source, "TEST_TIME", "time.Time", ctx)

	// Special types
	testGet[string](source, "TEST_URL", "string(url)", ctx)
	testGet[uuid.UUID](source, "TEST_UUID", "uuid.UUID", ctx)
	testGet[string](source, "TEST_IP", "string(ip)", ctx)
	testGet[os.FileMode](source, "TEST_FILEMODE", "os.FileMode", ctx)
	testGet[string](source, "TEST_PATH", "string(path)", ctx)

	// Slice types
	testGet[[]string](source, "TEST_STRING_SLICE", "[]string", ctx)
	testGet[[]int64](source, "TEST_INT64_SLICE", "[]int64", ctx)
	testGet[[]int](source, "TEST_INT_SLICE", "[]int", ctx)
	testGet[[]float64](source, "TEST_FLOAT64_SLICE", "[]float64", ctx)
	testGet[[]bool](source, "TEST_BOOL_SLICE", "[]bool", ctx)
}

// defineAllTypes adds all type definitions inside a command config callback
func defineAllTypes(cc *commandkit.CommandConfig, withEnv, withFlag, withFile bool) {
	d := func(key string) *commandkit.DefinitionBuilder {
		b := cc.Define(key)
		if withEnv {
			b.Env(key)
		}
		if withFlag {
			b.Flag(strings.ToLower(key))
		}
		if withFile {
			b.File(strings.ToLower(key))
		}
		return b
	}

	// Basic types
	d("TEST_STRING").String().Default("default_string")
	d("TEST_BOOL").Bool().Default(true)

	// Integer types
	d("TEST_INT64").Int64().Default(1)
	d("TEST_INT").Int().Default(2)
	d("TEST_UINT").Uint().Default(3)
	d("TEST_UINT8").Uint8().Default(4)
	d("TEST_UINT16").Uint16().Default(5)
	d("TEST_UINT32").Uint32().Default(6)
	d("TEST_UINT64").Uint64().Default(7)

	// Float types
	d("TEST_FLOAT64").Float64().Default(1.1)
	d("TEST_FLOAT32").Float32().Default(2.2)

	// Time types
	d("TEST_DURATION").Duration().Default(time.Minute)
	d("TEST_TIME").Time().Default(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	// Special types
	d("TEST_URL").URL().Default("https://default.example.com")
	d("TEST_UUID").UUID().Default(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	d("TEST_IP").IP().Default("127.0.0.1")
	d("TEST_FILEMODE").FileMode().Default(os.FileMode(0644))
	d("TEST_PATH").Path().Default("/default/path")

	// Slice types
	d("TEST_STRING_SLICE").StringSlice().Default([]string{"d1", "d2"})
	d("TEST_INT64_SLICE").Int64Slice().Default([]int64{1, 2})
	d("TEST_INT_SLICE").IntSlice().Default([]int{3, 4})
	d("TEST_FLOAT64_SLICE").Float64Slice().Default([]float64{1.1, 2.2})
	d("TEST_BOOL_SLICE").BoolSlice().Default([]bool{true, false})
}

// testSourceWithCommand creates a config, adds a command that runs the tests, and executes it
func testSourceWithCommand(source string, withEnv, withFlag, withFile bool, extraArgs []string) {
	cfg := commandkit.New()

	if withFile {
		err := cfg.LoadFile(testConfigPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load test config file: %v\n", err)
		}
	}

	cfg.Command("test").
		Func(func(ctx *commandkit.CommandContext) error {
			runAllTypeTests(source, ctx)
			return nil
		}).
		ShortHelp("Run conversion tests").
		Config(func(cc *commandkit.CommandConfig) {
			defineAllTypes(cc, withEnv, withFlag, withFile)
		})

	args := append([]string{"conversions", "test"}, extraArgs...)
	err := cfg.Execute(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing %s source test: %v\n", source, err)
	}
}

// --- Environment variable helpers ---

func setupEnvVars() {
	os.Setenv("TEST_STRING", "env_string_value")
	os.Setenv("TEST_BOOL", "false")
	os.Setenv("TEST_INT64", "42")
	os.Setenv("TEST_INT", "21")
	os.Setenv("TEST_UINT", "10")
	os.Setenv("TEST_UINT8", "255")
	os.Setenv("TEST_UINT16", "65535")
	os.Setenv("TEST_UINT32", "4294967295")
	os.Setenv("TEST_UINT64", "18446744073709551615")
	os.Setenv("TEST_FLOAT64", "3.14159")
	os.Setenv("TEST_FLOAT32", "2.718")
	os.Setenv("TEST_DURATION", "5m")
	os.Setenv("TEST_TIME", "2023-12-25T10:00:00Z")
	os.Setenv("TEST_URL", "https://env.example.com")
	os.Setenv("TEST_UUID", "550e8400-e29b-41d4-a716-446655440000")
	os.Setenv("TEST_IP", "192.168.1.1")
	os.Setenv("TEST_FILEMODE", "0755")
	os.Setenv("TEST_PATH", "/tmp/env_path")
	os.Setenv("TEST_STRING_SLICE", "e1,e2,e3")
	os.Setenv("TEST_INT64_SLICE", "10,20,30")
	os.Setenv("TEST_INT_SLICE", "40,50,60")
	os.Setenv("TEST_FLOAT64_SLICE", "4.4,5.5,6.6")
	os.Setenv("TEST_BOOL_SLICE", "false,true,false")
}

func clearEnvVars() {
	keys := []string{
		"TEST_STRING", "TEST_BOOL", "TEST_INT64", "TEST_INT",
		"TEST_UINT", "TEST_UINT8", "TEST_UINT16", "TEST_UINT32", "TEST_UINT64",
		"TEST_FLOAT64", "TEST_FLOAT32", "TEST_DURATION", "TEST_TIME",
		"TEST_URL", "TEST_UUID", "TEST_IP", "TEST_FILEMODE", "TEST_PATH",
		"TEST_STRING_SLICE", "TEST_INT64_SLICE", "TEST_INT_SLICE",
		"TEST_FLOAT64_SLICE", "TEST_BOOL_SLICE",
	}
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

// --- File config helpers ---

func testConfigPath() string {
	return "/tmp/commandkit_test_config.json"
}

func setupFileConfig() {
	data := map[string]any{
		"test_string":        "file_string_value",
		"test_bool":          "false",
		"test_int64":         "84",
		"test_int":           "42",
		"test_uint":          "20",
		"test_uint8":         "128",
		"test_uint16":        "32768",
		"test_uint32":        "2147483648",
		"test_uint64":        "9223372036854775808",
		"test_float64":       "6.28318",
		"test_float32":       "1.414",
		"test_duration":      "10m",
		"test_time":          "2025-06-15T12:00:00Z",
		"test_url":           "https://file.example.com",
		"test_uuid":          "550e8400-e29b-41d4-a716-446655440001",
		"test_ip":            "10.0.0.1",
		"test_filemode":      "0700",
		"test_path":          "/var/log/file_path",
		"test_string_slice":  "f1,f2,f3",
		"test_int64_slice":   "100,200,300",
		"test_int_slice":     "400,500,600",
		"test_float64_slice": "7.7,8.8,9.9",
		"test_bool_slice":    "true,true,false",
	}
	b, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(testConfigPath(), b, 0644)
}

func cleanupFileConfig() {
	os.Remove(testConfigPath())
}

// --- Flag args builder ---

func flagArgs() []string {
	return []string{
		"--test_string", "flag_string_value",
		"--test_bool", "false",
		"--test_int64", "100",
		"--test_int", "50",
		"--test_uint", "30",
		"--test_uint8", "200",
		"--test_uint16", "40000",
		"--test_uint32", "3000000000",
		"--test_uint64", "10000000000000000000",
		"--test_float64", "9.8696",
		"--test_float32", "1.732",
		"--test_duration", "15m",
		"--test_time", "2026-01-01T00:00:00Z",
		"--test_url", "https://flag.example.com",
		"--test_uuid", "550e8400-e29b-41d4-a716-446655440002",
		"--test_ip", "172.16.0.1",
		"--test_filemode", "0600",
		"--test_path", "/opt/flag_path",
		"--test_string_slice", "g1,g2,g3",
		"--test_int64_slice", "1000,2000,3000",
		"--test_int_slice", "4000,5000,6000",
		"--test_float64_slice", "10.1,20.2,30.3",
		"--test_bool_slice", "true,false,true",
	}
}

// --- Output ---

func printResults() {
	fmt.Printf("\n%-7s | %-25s | %-15s | %-15s | %-20s | %s\n",
		"Source", "Key", "Expected Type", "Returned Type", "Value", "OK")
	fmt.Println(strings.Repeat("-", 110))

	for _, r := range allResults {
		ok := "✅"
		errInfo := ""
		if !r.Success {
			ok = "❌"
			errInfo = r.Error
			if len(errInfo) > 40 {
				errInfo = errInfo[:37] + "..."
			}
		}
		val := r.Value
		if len(val) > 18 {
			val = val[:15] + "..."
		}
		if errInfo != "" {
			fmt.Printf("%-7s | %-25s | %-15s | %-15s | %-20s | %s %s\n",
				r.Source, r.Key, r.ExpectedType, r.ReturnedType, val, ok, errInfo)
		} else {
			fmt.Printf("%-7s | %-25s | %-15s | %-15s | %-20s | %s\n",
				r.Source, r.Key, r.ExpectedType, r.ReturnedType, val, ok)
		}
	}
}

func printSummary() {
	total := len(allResults)
	passed := 0
	failed := 0
	for _, r := range allResults {
		if r.Success {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total:  %d\n", total)
	fmt.Printf("Passed: %d ✅\n", passed)
	fmt.Printf("Failed: %d ❌\n", failed)

	if failed > 0 {
		fmt.Printf("\nFailed tests:\n")
		for _, r := range allResults {
			if !r.Success {
				fmt.Printf("  %-7s %-25s %s\n", r.Source, r.Key, r.Error)
			}
		}
	}
}

// --- Main ---

func main() {
	fmt.Println("=== CommandKit Type Conversion TDD Test Suite ===")
	fmt.Println("Testing all type conversions across all sources (DEFAULT, ENV, FLAG, FILE)")

	// 1. DEFAULT source (no env, no flags, no file)
	clearEnvVars()
	testSourceWithCommand("DEFAULT", false, false, false, nil)

	// 2. ENV source
	setupEnvVars()
	testSourceWithCommand("ENV", true, false, false, nil)
	clearEnvVars()

	// 3. FLAG source
	testSourceWithCommand("FLAG", false, true, false, flagArgs())

	// 4. FILE source
	setupFileConfig()
	defer cleanupFileConfig()
	testSourceWithCommand("FILE", false, false, true, nil)

	// Output
	printResults()
	printSummary()

	// Exit code
	for _, r := range allResults {
		if !r.Success {
			os.Exit(1)
		}
	}
}
