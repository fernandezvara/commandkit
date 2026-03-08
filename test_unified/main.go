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
	_, err := commandkit.Get[string](ctx, "DATABASE_URL")
	if err != nil {
		fmt.Printf("✅ Expected error caught: %v\n", err)
	} else {
		fmt.Printf("❌ Expected error but got success\n")
	}

	// Test 2: Config.Execute() unified entry point
	if err := cfg.Execute([]string{"test"}); err != nil {
		fmt.Printf("✅ Config.Execute() error caught: %v\n", err)
	} else {
		fmt.Printf("❌ Expected Config.Execute() error but got success\n")
	}

	fmt.Println("Unified error handling test completed!")
}
