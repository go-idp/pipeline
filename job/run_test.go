package job

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-idp/pipeline/step"
)

func TestJobTimeout(t *testing.T) {
	t.Run("job should timeout when timeout is exceeded", func(t *testing.T) {
		job := &Job{
			Name:    "test job timeout",
			Timeout: 1, // 1 second timeout
			Steps: []*step.Step{
				{
					Name:    "sleep step",
					Command: "sleep 3", // Sleep for 3 seconds, should timeout
				},
			},
		}

		if err := job.Setup("test-job-timeout"); err != nil {
			t.Fatalf("Failed to setup job: %v", err)
		}

		err := job.Run(context.Background())
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Check if error is timeout related (context.DeadlineExceeded or wrapped error)
		if !strings.Contains(err.Error(), "timeout") && 
		   !strings.Contains(err.Error(), "deadline exceeded") && 
		   err != context.DeadlineExceeded {
			t.Errorf("Expected timeout error, got: %v", err)
		}

		// Check job state
		if job.State == nil {
			t.Fatal("Job state is nil")
		}

		if job.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", job.State.Status)
		}

		if job.State.Error == "" {
			t.Error("Expected error message, but got empty string")
		}

		// Check if error message contains timeout information
		if !strings.Contains(job.State.Error, "timeout") && 
		   !strings.Contains(job.State.Error, "deadline exceeded") {
			t.Errorf("Expected timeout in error message, got: %s", job.State.Error)
		}

		if job.State.FailedAt.IsZero() {
			t.Error("Expected FailedAt to be set")
		}
	})

	t.Run("job should succeed when timeout is not exceeded", func(t *testing.T) {
		job := &Job{
			Name:    "test job success",
			Timeout: 10, // 10 seconds timeout
			Steps: []*step.Step{
				{
					Name:    "quick step",
					Command: "echo 'hello'", // Quick command, should succeed
				},
			},
		}

		if err := job.Setup("test-job-success"); err != nil {
			t.Fatalf("Failed to setup job: %v", err)
		}

		err := job.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Check job state
		if job.State == nil {
			t.Fatal("Job state is nil")
		}

		if job.State.Status != "succeeded" {
			t.Errorf("Expected status 'succeeded', got '%s'", job.State.Status)
		}

		if job.State.SucceedAt.IsZero() {
			t.Error("Expected SucceedAt to be set")
		}
	})

	t.Run("job should timeout on second step if timeout is exceeded", func(t *testing.T) {
		job := &Job{
			Name:    "test job timeout on second step",
			Timeout: 2, // 2 seconds timeout
			Steps: []*step.Step{
				{
					Name:    "quick step 1",
					Command: "echo 'step 1'", // Quick command
				},
				{
					Name:    "sleep step",
					Command: "sleep 3", // Sleep for 3 seconds, should timeout
				},
			},
		}

		if err := job.Setup("test-job-timeout-second-step"); err != nil {
			t.Fatalf("Failed to setup job: %v", err)
		}

		err := job.Run(context.Background())
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Check job state
		if job.State == nil {
			t.Fatal("Job state is nil")
		}

		if job.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", job.State.Status)
		}
	})
}

func TestJobTimeoutWithContext(t *testing.T) {
	t.Run("job should respect parent context timeout", func(t *testing.T) {
		// Create a parent context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		job := &Job{
			Name:    "test job with parent context",
			Timeout: 10, // Job timeout is longer, but parent context should take precedence
			Steps: []*step.Step{
				{
					Name:    "sleep step",
					Command: "sleep 2", // Should timeout due to parent context
				},
			},
		}

		if err := job.Setup("test-job-parent-context"); err != nil {
			t.Fatalf("Failed to setup job: %v", err)
		}

		err := job.Run(ctx)
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Should fail due to parent context timeout
		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Logf("Got error: %v", err)
		}
	})
}
