package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/stage"
	"github.com/go-idp/pipeline/step"
)

func TestPipelineTimeout(t *testing.T) {
	t.Run("pipeline should timeout when timeout is exceeded", func(t *testing.T) {
		pipeline := &Pipeline{
			Name:    "test pipeline timeout",
			Timeout: 1, // 1 second timeout
			Stages: []*stage.Stage{
				{
					Name: "test stage",
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
				},
			},
		}

		err := pipeline.Run(context.Background())
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Check if error is timeout related (context.DeadlineExceeded or wrapped error)
		if !strings.Contains(err.Error(), "timeout") && 
		   !strings.Contains(err.Error(), "deadline exceeded") && 
		   err != context.DeadlineExceeded {
			t.Errorf("Expected timeout error, got: %v", err)
		}

		// Check pipeline state
		if pipeline.State == nil {
			t.Fatal("Pipeline state is nil")
		}

		if pipeline.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", pipeline.State.Status)
		}

		if pipeline.State.Error == "" {
			t.Error("Expected error message, but got empty string")
		}

		// Check if error message contains timeout information
		if !strings.Contains(pipeline.State.Error, "timeout") && 
		   !strings.Contains(pipeline.State.Error, "deadline exceeded") {
			t.Errorf("Expected timeout in error message, got: %s", pipeline.State.Error)
		}

		if pipeline.State.FailedAt.IsZero() {
			t.Error("Expected FailedAt to be set")
		}
	})

	t.Run("pipeline should succeed when timeout is not exceeded", func(t *testing.T) {
		pipeline := &Pipeline{
			Name:    "test pipeline success",
			Timeout: 10, // 10 seconds timeout
			Stages: []*stage.Stage{
				{
					Name: "test stage",
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
				},
			},
		}

		err := pipeline.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Check pipeline state
		if pipeline.State == nil {
			t.Fatal("Pipeline state is nil")
		}

		if pipeline.State.Status != "succeeded" {
			t.Errorf("Expected status 'succeeded', got '%s'", pipeline.State.Status)
		}

		if pipeline.State.SucceedAt.IsZero() {
			t.Error("Expected SucceedAt to be set")
		}
	})

	t.Run("pipeline should use default timeout when timeout is 0", func(t *testing.T) {
		pipeline := &Pipeline{
			Name:    "test pipeline default timeout",
			Timeout: 0, // Should use default
			Stages: []*stage.Stage{
				{
					Name: "test stage",
					Jobs: []*job.Job{
						{
							Name: "test job",
							Steps: []*step.Step{
								{
									Name:    "quick step",
									Command: "echo 'hello'",
								},
							},
						},
					},
				},
			},
		}

		err := pipeline.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// After prepare, timeout should be set to default
		if pipeline.Timeout == 0 {
			t.Error("Expected timeout to be set to default value")
		}
	})
}

func TestPipelineTimeoutWithContext(t *testing.T) {
	t.Run("pipeline should respect parent context timeout", func(t *testing.T) {
		// Create a parent context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		pipeline := &Pipeline{
			Name:    "test pipeline with parent context",
			Timeout: 10, // Pipeline timeout is longer, but parent context should take precedence
			Stages: []*stage.Stage{
				{
					Name: "test stage",
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
				},
			},
		}

		err := pipeline.Run(ctx)
		if err == nil {
			t.Fatal("Expected timeout error, but got nil")
		}

		// Should fail due to parent context timeout
		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Logf("Got error: %v", err)
		}
	})
}
