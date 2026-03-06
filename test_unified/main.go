package main

import (
	"fmt"

	"github.com/fernandezvara/commandkit"
)

func main() {
	cfg := commandkit.New()

	// Define a required config
	cfg.Define("DATABASE_URL").
		String().
		Env("DATABASE_URL").
		Required().
		Secret().
		Description("Database connection string")

	// Create command context
	ctx := commandkit.NewCommandContext([]string{}, cfg, "test", "")

	// Test the new Get API with CommandResult
	fmt.Println("Testing unified error handling...")

	// Test 1: Missing required value
	result := commandkit.Get[string](ctx, "DATABASE_URL")
	if result.Error != nil {
		fmt.Printf("✅ Expected error caught: %v\n", result.Error)
		fmt.Printf("   Message: %s\n", result.Message)
		fmt.Printf("   ShouldExit: %v\n", result.ShouldExit)
	} else {
		fmt.Printf("❌ Expected error but got success\n")
	}

	// Test 2: Config.Process() with CommandResult
	processResult := cfg.Process()
	if processResult.Error != nil {
		fmt.Printf("✅ Config.Process() error caught: %v\n", processResult.Error)
		fmt.Printf("   Message: %s\n", processResult.Message)
		fmt.Printf("   ShouldExit: %v\n", processResult.ShouldExit)
	} else {
		fmt.Printf("❌ Expected Config.Process() error but got success\n")
	}

	fmt.Println("Unified error handling test completed!")
}
