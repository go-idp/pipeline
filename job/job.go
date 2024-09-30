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
			Image:       j.Image,
			Environment: j.Environment,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Run runs steps in sequence
func (j *Job) Run(ctx context.Context) error {
	j.logger.Infof("[job: %s] start", j.Name)
	defer j.logger.Infof("[job: %s] done", j.Name)

	for _, step := range j.Steps {
		if err := step.Run(ctx); err != nil {
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
