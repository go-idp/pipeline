package server

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-idp/pipeline"
	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/stage"
	"github.com/go-idp/pipeline/step"
)

func testPipeline(name string) *pipeline.Pipeline {
	return &pipeline.Pipeline{
		Name: name,
		Stages: []*stage.Stage{
			{
				Name: "s",
				Jobs: []*job.Job{
					{
						Name: "j",
						Steps: []*step.Step{
							{Name: "hello", Command: "echo hello"},
						},
					},
				},
			},
		},
	}
}

func TestWithRecovery(t *testing.T) {
	// panic -> error, not a crash
	err := withRecovery(func() error {
		panic("boom")
	})
	if err == nil || !strings.Contains(err.Error(), "panic: boom") {
		t.Fatalf("expected panic converted to error, got: %v", err)
	}

	// normal error passthrough
	inner := errors.New("normal")
	if err := withRecovery(func() error { return inner }); err != inner {
		t.Fatalf("expected original error, got: %v", err)
	}

	// success
	if err := withRecovery(func() error { return nil }); err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}
}

func TestTaskExecutor_DefaultTimeoutInjected(t *testing.T) {
	var out bytes.Buffer

	pl := testPipeline("timeout-test")
	executor := &taskExecutor{
		workdir:     t.TempDir(),
		executor:    TaskExecutorInProcess,
		timeout:     3600,
		environment: map[string]string{},
	}

	if err := executor.Run("id-1", pl, context.Background(), &out, &out); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if pl.Timeout != 3600 {
		t.Errorf("default timeout should be injected, got %d", pl.Timeout)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Errorf("expected step output in log, got: %q", out.String())
	}
}

func TestTaskExecutor_ExplicitTimeoutWins(t *testing.T) {
	pl := testPipeline("timeout-explicit")
	pl.Timeout = 12345

	executor := &taskExecutor{
		workdir:     t.TempDir(),
		executor:    TaskExecutorInProcess,
		timeout:     3600,
		environment: map[string]string{},
	}

	if err := executor.Run("id-1", pl, context.Background(), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if pl.Timeout != 12345 {
		t.Errorf("explicit timeout should win, got %d", pl.Timeout)
	}
}

func TestTaskExecutor_SubprocessSuccess(t *testing.T) {
	workdir := t.TempDir()

	pl := testPipeline("subprocess-ok")
	executor := &taskExecutor{
		workdir:     workdir,
		executor:    TaskExecutorSubprocess,
		timeout:     3600,
		environment: map[string]string{"CI": "true"},
		executable:  "/usr/bin/true", // exits 0, ignores args
	}

	var out bytes.Buffer
	if err := executor.Run("id-1", pl, context.Background(), &out, &out); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// the yaml should be written with the embedded workdir and default timeout
	data, err := os.ReadFile(filepath.Join(workdir, "id-1", "pipeline.yaml"))
	if err != nil {
		t.Fatalf("pipeline.yaml not written: %v", err)
	}
	if !strings.Contains(string(data), "timeout: 3600") {
		t.Errorf("pipeline.yaml should contain the default timeout, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), filepath.Join(workdir, "id-1")) {
		t.Errorf("pipeline.yaml should contain the workdir, got:\n%s", string(data))
	}
}

func TestTaskExecutor_SubprocessFailure(t *testing.T) {
	executor := &taskExecutor{
		workdir:    t.TempDir(),
		executor:   TaskExecutorSubprocess,
		executable: "/usr/bin/false", // exits 1
	}

	err := executor.Run("id-1", testPipeline("subprocess-fail"), context.Background(), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "failed to run pipeline in subprocess") {
		t.Fatalf("expected subprocess failure error, got: %v", err)
	}
}

func TestTaskExecutor_SubprocessCancelled(t *testing.T) {
	// a script that sleeps; cancellation must kill it (and its process group)
	script := filepath.Join(t.TempDir(), "slow.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30\necho done\n"), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	executor := &taskExecutor{
		workdir:    t.TempDir(),
		executor:   TaskExecutorSubprocess,
		executable: script,
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- executor.Run("id-1", testPipeline("subprocess-cancel"), ctx, &bytes.Buffer{}, &bytes.Buffer{})
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("subprocess did not terminate after cancellation")
	}
}

func TestEnvPairs(t *testing.T) {
	pairs := envPairs(map[string]string{"A": "1", "B": "2"})
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %v", pairs)
	}

	joined := strings.Join(pairs, " ")
	for _, want := range []string{"A=1", "B=2"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %v", want, pairs)
		}
	}
}

func TestTaskExecutor_SubprocessEnvCarried(t *testing.T) {
	workdir := t.TempDir()

	// a script that prints the environment; verify the whitelisted var is passed
	envScript := filepath.Join(workdir, "env.sh")
	if err := os.WriteFile(envScript, []byte("#!/bin/sh\nenv\n"), 0o755); err != nil {
		t.Fatalf("failed to write env script: %v", err)
	}

	pl := testPipeline("subprocess-env")
	executor := &taskExecutor{
		workdir:     workdir,
		executor:    TaskExecutorSubprocess,
		timeout:     60,
		environment: map[string]string{"PIPELINE_TASK_EXECUTOR_TEST": "carried"},
		executable:  envScript,
	}

	var out bytes.Buffer
	if err := executor.Run("id-1", pl, context.Background(), &out, &out); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(out.String(), "PIPELINE_TASK_EXECUTOR_TEST=carried") {
		t.Errorf("whitelisted env should be passed to the subprocess, got:\n%s", out.String())
	}
}
