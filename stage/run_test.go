package stage

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/step"
)

func TestStageTimeout(t *testing.T) {
	t.Run("stage should timeout when timeout is exceeded", func(t *testing.T) {
		stage := &Stage{
			Name:    "test stage timeout",
			Timeout: 1, // 1 second timeout
			Jobs: []*job.Job{
				{
					Name: "test job",
					Steps: []*step.Step{
						{
							Name:    "sleep step",
							Command: "sleep 3", // Sleep for 3 seconds, should timeout
						},
					},
				},
			},
		}

		if err := stage.Setup("test-stage-timeout"); err != nil {
			t.Fatalf("Failed to setup stage: %v", err)
		}

		err := stage.Run(context.Background())
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Check if error is timeout related (context.DeadlineExceeded or wrapped error)
		if !strings.Contains(err.Error(), "timeout") &&
			!strings.Contains(err.Error(), "deadline exceeded") &&
			err != context.DeadlineExceeded {
			t.Errorf("Expected timeout error, got: %v", err)
		}

		// Check stage state
		if stage.State == nil {
			t.Fatal("Stage state is nil")
		}

		if stage.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", stage.State.Status)
		}

		if stage.State.Error == "" {
			t.Error("Expected error message, but got empty string")
		}

		// Check if error message contains timeout information
		if !strings.Contains(stage.State.Error, "timeout") &&
			!strings.Contains(stage.State.Error, "deadline exceeded") {
			t.Errorf("Expected timeout in error message, got: %s", stage.State.Error)
		}

		if stage.State.FailedAt.IsZero() {
			t.Error("Expected FailedAt to be set")
		}
	})

	t.Run("stage should succeed when timeout is not exceeded", func(t *testing.T) {
		stage := &Stage{
			Name:    "test stage success",
			Timeout: 10, // 10 seconds timeout
			Jobs: []*job.Job{
				{
					Name: "test job",
					Steps: []*step.Step{
						{
							Name:    "quick step",
							Command: "echo 'hello'", // Quick command, should succeed
						},
					},
				},
			},
		}

		if err := stage.Setup("test-stage-success"); err != nil {
			t.Fatalf("Failed to setup stage: %v", err)
		}

		err := stage.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Check stage state
		if stage.State == nil {
			t.Fatal("Stage state is nil")
		}

		if stage.State.Status != "succeeded" {
			t.Errorf("Expected status 'succeeded', got '%s'", stage.State.Status)
		}

		if stage.State.SucceedAt.IsZero() {
			t.Error("Expected SucceedAt to be set")
		}
	})

	t.Run("stage parallel mode should timeout correctly", func(t *testing.T) {
		stage := &Stage{
			Name:    "test stage parallel timeout",
			Timeout: 1, // 1 second timeout
			RunMode: RunModeParallel,
			Jobs: []*job.Job{
				{
					Name: "test job 1",
					Steps: []*step.Step{
						{
							Name:    "sleep step 1",
							Command: "sleep 3", // Should timeout
						},
					},
				},
				{
					Name: "test job 2",
					Steps: []*step.Step{
						{
							Name:    "sleep step 2",
							Command: "sleep 3", // Should timeout
						},
					},
				},
			},
		}

		if err := stage.Setup("test-stage-parallel-timeout"); err != nil {
			t.Fatalf("Failed to setup stage: %v", err)
		}

		err := stage.Run(context.Background())
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Check stage state
		if stage.State == nil {
			t.Fatal("Stage state is nil")
		}

		if stage.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", stage.State.Status)
		}
	})
}

func TestStageTimeoutWithContext(t *testing.T) {
	t.Run("stage should respect parent context timeout", func(t *testing.T) {
		// Create a parent context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		stage := &Stage{
			Name:    "test stage with parent context",
			Timeout: 10, // Stage timeout is longer, but parent context should take precedence
			Jobs: []*job.Job{
				{
					Name: "test job",
					Steps: []*step.Step{
						{
							Name:    "sleep step",
							Command: "sleep 2", // Should timeout due to parent context
						},
					},
				},
			},
		}

		if err := stage.Setup("test-stage-parent-context"); err != nil {
			t.Fatalf("Failed to setup stage: %v", err)
		}

		err := stage.Run(ctx)
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Should fail due to parent context timeout
		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Logf("Got error: %v", err)
		}
	})
}
