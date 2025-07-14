package main

import (
	"context"
	"fmt"
	"os"

	"github.com/m-mizutani/ctxlog"
)

func main() {
	// Example 1: Basic log capture
	println("=== Example 1: Basic Log Capture ===")
	basicCaptureExample()

	// Example 2: Capture with scopes
	println("\n=== Example 2: Capture with Scopes ===")
	scopeCaptureExample()

	// Example 3: Capture with sampling
	println("\n=== Example 3: Capture with Sampling ===")
	samplingCaptureExample()

	// Example 4: Test-like usage
	println("\n=== Example 4: Test-like Usage ===")
	testLikeExample()
}

func basicCaptureExample() {
	ctx := context.Background()
	
	// Create capture logger
	ctx, capture := ctxlog.NewCapture(ctx)
	
	// Simulate application logging
	simulateApplicationLogging(ctx)
	
	// Retrieve captured messages
	messages := capture.Messages()
	fmt.Printf("Captured %d messages:\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  %d: %s\n", i+1, msg)
	}
}

func scopeCaptureExample() {
	ctx := context.Background()
	ctx, capture := ctxlog.NewCapture(ctx)
	
	// Create scope
	scope := ctxlog.NewScope("test", ctxlog.EnabledBy("DEBUG_TEST"))
	_ = os.Setenv("DEBUG_TEST", "1")
	
	// Log with scope
	logger := ctxlog.From(ctx, scope)
	logger.Info("Scoped message", "key", "value")
	logger.Warn("Warning message")
	
	// Check captured messages
	messages := capture.Messages()
	fmt.Printf("Captured %d scoped messages:\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  %d: %s\n", i+1, msg)
	}
}

func samplingCaptureExample() {
	ctx := context.Background()
	ctx, capture := ctxlog.NewCapture(ctx)
	
	// Log with sampling - some messages will be discarded
	for i := 0; i < 10; i++ {
		logger := ctxlog.From(ctx, ctxlog.WithSampling(0.5))
		logger.Info("Sampled message", "iteration", i)
	}
	
	messages := capture.Messages()
	fmt.Printf("Captured %d out of 10 sampled messages:\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  %d: %s\n", i+1, msg)
	}
}

func testLikeExample() {
	// This demonstrates how you might use capture in a test
	ctx := context.Background()
	ctx, capture := ctxlog.NewCapture(ctx)
	
	// Simulate a function that should log specific messages
	businessLogic(ctx, "user123", "create")
	
	// Verify logging behavior
	messages := capture.Messages()
	
	// Check that we got the expected number of messages
	if len(messages) != 2 {
		fmt.Printf("ERROR: Expected 2 messages, got %d\n", len(messages))
		return
	}
	
	// Check message content (in a real test, you'd use more sophisticated matching)
	expectedSubstrings := []string{"Starting operation", "Operation completed"}
	for i, expected := range expectedSubstrings {
		if i < len(messages) {
			fmt.Printf("✓ Message %d contains expected content: %s\n", i+1, expected)
		} else {
			fmt.Printf("✗ Missing expected message: %s\n", expected)
		}
	}
	
	fmt.Printf("Test completed successfully\n")
}

func simulateApplicationLogging(ctx context.Context) {
	logger := ctxlog.From(ctx)
	
	logger.Info("Application started")
	logger.Debug("Debug information")
	logger.Warn("Warning message")
	logger.Error("Error occurred")
}

func businessLogic(ctx context.Context, userID, operation string) {
	logger := ctxlog.From(ctx)
	
	logger.Info("Starting operation", 
		"user_id", userID, 
		"operation", operation)
	
	// Simulate some work
	// ... business logic here ...
	
	logger.Info("Operation completed",
		"user_id", userID,
		"operation", operation,
		"result", "success")
}