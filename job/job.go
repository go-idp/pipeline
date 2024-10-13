package job

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-idp/pipeline/step"
	"github.com/go-zoox/logger"
)

type Job struct {
	Name  string       `json:"name" yaml:"name"`
	Steps []*step.Step `json:"steps" yaml:"steps"`
	//
	Workdir string `json:"workdir" yaml:"workdir"`
	//
	Image       string            `json:"image" yaml:"image"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	//
	State *State `json:"state" yaml:"state"`
	//
	stdout io.Writer
	stderr io.Writer
	//
	logger *logger.Logger
}

type RunConfig struct {
	Total   int
	Current int
	//
	Parent string
}

type RunOption func(cfg *RunConfig)

func (s *Job) getLogger() *logger.Logger {
	l := logger.New()
	l.SetStdout(s.stdout)
	return l
}

// Setup sets up the job
func (j *Job) Setup(id string, opts ...*Job) error {
	if j.stdout == nil {
		j.stdout = os.Stdout

		if j.stderr == nil {
			j.stderr = j.stdout
		}
	}

	j.logger = j.getLogger()

	// merge config
	for _, opt := range opts {
		if j.Image == "" {
			j.Image = opt.Image
		}

		if j.Workdir == "" {
			j.Workdir = opt.Workdir
		}

		if j.Environment == nil {
			j.Environment = opt.Environment
		} else {
			for k, v := range opt.Environment {
				if _, ok := j.Environment[k]; !ok {
					j.Environment[k] = v
				}
			}
		}
	}

	// setup state
	j.State = &State{
		ID:     id,
		Status: "running",
		//
		StartedAt: time.Now(),
	}

	// setup steps
	for index, s := range j.Steps {
		err := s.Setup(fmt.Sprintf("%s.%d", j.State.ID, index), &step.Step{
			Workdir: j.Workdir,
			//
			Environment: j.Environment,
			//
			Image: j.Image,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Run runs steps in sequence
func (j *Job) Run(ctx context.Context, opts ...RunOption) error {
	cfg := &RunConfig{}
	for _, o := range opts {
		o(cfg)
	}

	j.logger.Infof("%s[job(%d/%d): %s] start", cfg.Parent, cfg.Current, cfg.Total, j.Name)
	defer j.logger.Infof("%s[job(%d/%d): %s] done", cfg.Parent, cfg.Current, cfg.Total, j.Name)

	for i, s := range j.Steps {
		err := s.Run(ctx, func(c *step.RunConfig) {
			c.Total = len(j.Steps)
			c.Current = i + 1
			c.Parent = fmt.Sprintf("%s[job(%d/%d): %s]", cfg.Parent, cfg.Current, cfg.Total, j.Name)
		})

		if err != nil {
			j.State.Status = "failed"
			j.State.Error = err.Error()
			j.State.FailedAt = time.Now()
			return err
		}
	}

	j.State.Status = "succeeded"
	j.State.SucceedAt = time.Now()

	return nil
}

func (j *Job) SetStdout(stdout io.Writer) {
	for _, step := range j.Steps {
		step.SetStdout(stdout)
	}
}

func (j *Job) SetStderr(stderr io.Writer) {
	for _, step := range j.Steps {
		step.SetStderr(stderr)
	}
}
