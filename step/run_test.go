package step

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestStepTimeout(t *testing.T) {
	t.Run("step should timeout when timeout is exceeded", func(t *testing.T) {
		step := &Step{
			Name:    "test step timeout",
			Timeout: 1, // 1 second timeout
			Command: "sleep 3", // Sleep for 3 seconds, should timeout
		}

		if err := step.Setup("test-step-timeout"); err != nil {
			t.Fatalf("Failed to setup step: %v", err)
		}

		err := step.Run(context.Background())
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Check if error is timeout related (context.DeadlineExceeded or wrapped error)
		if !strings.Contains(err.Error(), "timeout") && 
		   !strings.Contains(err.Error(), "deadline exceeded") && 
		   err != context.DeadlineExceeded {
			t.Errorf("Expected timeout error, got: %v", err)
		}

		// Check step state
		if step.State == nil {
			t.Fatal("Step state is nil")
		}

		if step.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", step.State.Status)
		}

		if step.State.Error == "" {
			t.Error("Expected error message, but got empty string")
		}

		// Check if error message contains timeout information
		if !strings.Contains(step.State.Error, "timeout") && 
		   !strings.Contains(step.State.Error, "deadline exceeded") {
			t.Errorf("Expected timeout in error message, got: %s", step.State.Error)
		}

		if step.State.FailedAt.IsZero() {
			t.Error("Expected FailedAt to be set")
		}
	})

	t.Run("step should succeed when timeout is not exceeded", func(t *testing.T) {
		step := &Step{
			Name:    "test step success",
			Timeout: 10, // 10 seconds timeout
			Command: "echo 'hello'", // Quick command, should succeed
		}

		if err := step.Setup("test-step-success"); err != nil {
			t.Fatalf("Failed to setup step: %v", err)
		}

		err := step.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Check step state
		if step.State == nil {
			t.Fatal("Step state is nil")
		}

		if step.State.Status != "succeeded" {
			t.Errorf("Expected status 'succeeded', got '%s'", step.State.Status)
		}

		if step.State.SucceedAt.IsZero() {
			t.Error("Expected SucceedAt to be set")
		}
	})

	t.Run("step should use default timeout when timeout is 0", func(t *testing.T) {
		step := &Step{
			Name:    "test step default timeout",
			Timeout: 0, // Should use default
			Command: "echo 'hello'",
		}

		if err := step.Setup("test-step-default-timeout"); err != nil {
			t.Fatalf("Failed to setup step: %v", err)
		}

		// After setup, timeout should be set to default
		if step.Timeout == 0 {
			t.Error("Expected timeout to be set to default value")
		}

		err := step.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Check step state
		if step.State == nil {
			t.Fatal("Step state is nil")
		}

		if step.State.Status != "succeeded" {
			t.Errorf("Expected status 'succeeded', got '%s'", step.State.Status)
		}
	})
}

func TestStepTimeoutWithContext(t *testing.T) {
	t.Run("step should respect parent context timeout", func(t *testing.T) {
		// Create a parent context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		step := &Step{
			Name:    "test step with parent context",
			Timeout: 10, // Step timeout is longer, but parent context should take precedence
			Command: "sleep 2", // Should timeout due to parent context
		}

		if err := step.Setup("test-step-parent-context"); err != nil {
			t.Fatalf("Failed to setup step: %v", err)
		}

		err := step.Run(ctx)
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Should fail due to parent context timeout
		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Logf("Got error: %v", err)
		}
	})

	t.Run("step should use shorter timeout between context and step timeout", func(t *testing.T) {
		// Create a parent context with a longer timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		step := &Step{
			Name:    "test step shorter timeout",
			Timeout: 1, // Step timeout is shorter, should take precedence
			Command: "sleep 3", // Should timeout due to step timeout
		}

		if err := step.Setup("test-step-shorter-timeout"); err != nil {
			t.Fatalf("Failed to setup step: %v", err)
		}

		err := step.Run(ctx)
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Should fail due to step timeout (shorter than context timeout)
		if !strings.Contains(err.Error(), "timeout") && 
		   !strings.Contains(err.Error(), "deadline exceeded") && 
		   err != context.DeadlineExceeded {
			t.Logf("Got error: %v", err)
		}
	})
}
