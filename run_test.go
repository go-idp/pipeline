package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/stage"
	"github.com/go-idp/pipeline/step"
	"github.com/go-zoox/fs"
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

// TestPipelineWorkdirCleanup tests that workdir is cleaned on success but preserved on failure
func TestPipelineWorkdirCleanup(t *testing.T) {
	t.Run("workdir should be cleaned when pipeline succeeds", func(t *testing.T) {
		// Create a temporary workdir
		tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir) // Cleanup in case of test failure

		workdir := filepath.Join(tmpDir, "workdir")
		if err := os.MkdirAll(workdir, 0755); err != nil {
			t.Fatalf("Failed to create workdir: %v", err)
		}

		// Create a test file in workdir to verify it gets cleaned
		testFile := filepath.Join(workdir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		pipeline := &Pipeline{
			Name:    "test pipeline success cleanup",
			Workdir: workdir,
			Timeout: 10,
			Stages: []*stage.Stage{
				{
					Name: "test stage",
					Jobs: []*job.Job{
						{
							Name: "test job",
							Steps: []*step.Step{
								{
									Name:    "success step",
									Command: "echo 'success'",
								},
							},
						},
					},
				},
			},
		}

		err = pipeline.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Verify workdir was cleaned
		if fs.IsExist(workdir) {
			t.Errorf("Expected workdir to be cleaned, but it still exists: %s", workdir)
		}

		// Verify pipeline state
		if pipeline.State.Status != "succeeded" {
			t.Errorf("Expected status 'succeeded', got '%s'", pipeline.State.Status)
		}
	})

	t.Run("workdir should be preserved when pipeline fails", func(t *testing.T) {
		// Create a temporary workdir
		tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir) // Cleanup after test

		workdir := filepath.Join(tmpDir, "workdir")
		if err := os.MkdirAll(workdir, 0755); err != nil {
			t.Fatalf("Failed to create workdir: %v", err)
		}

		// Create a test file in workdir to verify it's preserved
		testFile := filepath.Join(workdir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		pipeline := &Pipeline{
			Name:    "test pipeline failure preserve",
			Workdir: workdir,
			Timeout: 10,
			Stages: []*stage.Stage{
				{
					Name: "test stage",
					Jobs: []*job.Job{
						{
							Name: "test job",
							Steps: []*step.Step{
								{
									Name:    "fail step",
									Command: "exit 1", // This will fail
								},
							},
						},
					},
				},
			},
		}

		err = pipeline.Run(context.Background())
		if err == nil {
			t.Fatal("Expected error, but got nil")
		}

		// Verify workdir was preserved
		if !fs.IsExist(workdir) {
			t.Errorf("Expected workdir to be preserved, but it was cleaned: %s", workdir)
		}

		// Verify test file still exists
		if !fs.IsExist(testFile) {
			t.Errorf("Expected test file to be preserved, but it was deleted: %s", testFile)
		}

		// Verify pipeline state
		if pipeline.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", pipeline.State.Status)
		}

		if pipeline.State.Error == "" {
			t.Error("Expected error message, but got empty string")
		}
	})

	t.Run("workdir should not be cleaned if it's current directory", func(t *testing.T) {
		// Get current directory
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}

		pipeline := &Pipeline{
			Name:    "test pipeline current dir",
			Workdir: currentDir, // Use current directory
			Timeout: 10,
			Stages: []*stage.Stage{
				{
					Name: "test stage",
					Jobs: []*job.Job{
						{
							Name: "test job",
							Steps: []*step.Step{
								{
									Name:    "success step",
									Command: "echo 'success'",
								},
							},
						},
					},
				},
			},
		}

		err = pipeline.Run(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Verify current directory still exists (should not be cleaned)
		if !fs.IsExist(currentDir) {
			t.Errorf("Current directory should not be cleaned, but it doesn't exist")
		}

		// Verify pipeline state
		if pipeline.State.Status != "succeeded" {
			t.Errorf("Expected status 'succeeded', got '%s'", pipeline.State.Status)
		}
	})
}

// TestPipelineErrorHandling tests error handling and logging
func TestPipelineErrorHandling(t *testing.T) {
	t.Run("pipeline should set correct error state on failure", func(t *testing.T) {
		pipeline := &Pipeline{
			Name:    "test pipeline error handling",
			Timeout: 10,
			Stages: []*stage.Stage{
				{
					Name: "test stage",
					Jobs: []*job.Job{
						{
							Name: "test job",
							Steps: []*step.Step{
								{
									Name:    "fail step",
									Command: "exit 1",
								},
							},
						},
					},
				},
			},
		}

		err := pipeline.Run(context.Background())
		if err == nil {
			t.Fatal("Expected error, but got nil")
		}

		// Verify pipeline state
		if pipeline.State == nil {
			t.Fatal("Pipeline state is nil")
		}

		if pipeline.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", pipeline.State.Status)
		}

		if pipeline.State.Error == "" {
			t.Error("Expected error message, but got empty string")
		}

		if pipeline.State.FailedAt.IsZero() {
			t.Error("Expected FailedAt to be set")
		}

		if !strings.Contains(pipeline.State.Error, "exit status 1") {
			t.Logf("Error message: %s", pipeline.State.Error)
		}
	})

	t.Run("pipeline should handle timeout errors correctly", func(t *testing.T) {
		pipeline := &Pipeline{
			Name:    "test pipeline timeout error",
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
									Command: "sleep 3", // Should timeout
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

		// Verify pipeline state
		if pipeline.State.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", pipeline.State.Status)
		}

		// Check if error message contains timeout information
		if !strings.Contains(pipeline.State.Error, "timeout") &&
			!strings.Contains(pipeline.State.Error, "deadline exceeded") {
			t.Logf("Error message: %s", pipeline.State.Error)
		}
	})
}
