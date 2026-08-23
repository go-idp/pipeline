package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/encoding/yaml"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
)

// Task executors
const (
	// TaskExecutorInProcess runs the pipeline inside the server process
	// (goroutine). Default, fully compatible with the legacy behavior.
	TaskExecutorInProcess = "in-process"
	// TaskExecutorSubprocess runs the pipeline in a separate `pipeline run`
	// subprocess (reusing the running binary), isolating task crashes,
	// panics and resource usage from the server process.
	TaskExecutorSubprocess = "subprocess"
)

// taskExecutor runs a pipeline task with panic recovery, an optional default
// timeout, and optional subprocess isolation.
//
// It is server-only: the pipeline engine (`pipeline` package) and the `run`
// command are untouched.
type taskExecutor struct {
	// workdir is the server workdir; each task runs under workdir/<id>.
	workdir string
	// environment is the whitelisted environment applied to every task.
	environment map[string]string
	// executor is TaskExecutorInProcess (default) or TaskExecutorSubprocess.
	executor string
	// timeout is the default task timeout in seconds, applied when the
	// pipeline has no explicit timeout. 0 disables the default.
	timeout int64
	// executable overrides the binary used by the subprocess executor
	// (defaults to the current executable, i.e. `pipeline` itself).
	executable string
}

// Run executes the pipeline with panic recovery: a panicking task is reported
// as an error instead of crashing the server process.
func (t *taskExecutor) Run(id string, pl *pipeline.Pipeline, ctx context.Context, stdout, stderr io.Writer) (err error) {
	return withRecovery(func() error {
		// default task timeout: only applies when the pipeline has no explicit timeout
		if pl.Timeout == 0 && t.timeout > 0 {
			pl.Timeout = t.timeout
		}

		switch t.executor {
		case TaskExecutorSubprocess:
			return t.runSubprocess(id, pl, ctx, stdout, stderr)
		default:
			return t.runInProcess(id, pl, ctx, stdout, stderr)
		}
	})
}

// withRecovery converts a panic into an error so a single task cannot crash
// the whole server process.
func withRecovery(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("[task] panicked: %v", r)
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return fn()
}

// runInProcess runs the pipeline inside the current process (legacy behavior).
func (t *taskExecutor) runInProcess(id string, pl *pipeline.Pipeline, ctx context.Context, stdout, stderr io.Writer) error {
	pl.SetWorkdir(filepath.Join(t.workdir, id))
	pl.SetEnvironment(t.environment)
	if stdout != nil {
		pl.SetStdout(stdout)
	}
	if stderr != nil {
		pl.SetStderr(stderr)
	}

	return pl.Run(ctx, func(cfg *pipeline.RunConfig) {
		cfg.ID = id
	})
}

// runSubprocess runs the pipeline in a separate `pipeline run` subprocess,
// reusing the running pipeline binary. Task crashes/panics (including native
// service SDK panics) and resource usage stay inside the subprocess.
func (t *taskExecutor) runSubprocess(id string, pl *pipeline.Pipeline, ctx context.Context, stdout, stderr io.Writer) error {
	dir := filepath.Join(t.workdir, id)
	if err := fs.Mkdirp(dir); err != nil {
		return fmt.Errorf("failed to create task workdir: %s", err)
	}

	// embed the workdir and (default) timeout into the YAML so that
	// `pipeline run -c <file>` picks them up
	pl.SetWorkdir(dir)
	data, err := yaml.Encode(pl)
	if err != nil {
		return fmt.Errorf("failed to encode pipeline yaml: %s", err)
	}

	file := filepath.Join(dir, "pipeline.yaml")
	if err := fs.WriteFile(file, data); err != nil {
		return fmt.Errorf("failed to write pipeline yaml: %s", err)
	}

	exe := t.executable
	if exe == "" {
		exe, err = os.Executable()
		if err != nil {
			return fmt.Errorf("failed to locate pipeline executable: %s", err)
		}
	}

	cmd := exec.CommandContext(ctx, exe, "run", "-c", file)
	cmd.Env = append(os.Environ(), envPairs(t.environment)...)
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	// start the subprocess in its own process group and kill the whole group
	// on cancellation so step children are not orphaned
	configureProcessGroup(cmd)
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return killProcessGroup(cmd.Process.Pid)
	}

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("failed to run pipeline in subprocess: %s", err)
	}

	return nil
}

// envPairs flattens a map into KEY=VALUE strings.
func envPairs(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}
