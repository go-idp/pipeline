package stage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-idp/pipeline/job"
	"github.com/go-zoox/logger"
	"golang.org/x/sync/errgroup"
)

type Stage struct {
	Name string     `json:"name" yaml:"name"`
	Jobs []*job.Job `json:"jobs" yaml:"jobs"`
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

func (s *Stage) getLogger() *logger.Logger {
	l := logger.New()
	l.SetStdout(s.stdout)
	return l
}

// Setup sets up the stage
func (s *Stage) Setup(id string, opts ...*Stage) error {
	if s.stdout == nil {
		s.stdout = os.Stdout

		if s.stderr == nil {
			s.stderr = s.stdout
		}
	}

	s.logger = s.getLogger()

	// merge config
	for _, opt := range opts {
		if s.Image == "" {
			s.Image = opt.Image
		}

		if s.Workdir == "" {
			s.Workdir = opt.Workdir
		}

		if s.Environment == nil {
			s.Environment = opt.Environment
		} else {
			for k, v := range opt.Environment {
				if _, ok := s.Environment[k]; !ok {
					s.Environment[k] = v
				}
			}
		}
	}

	// setup state
	s.State = &State{
		ID:     id,
		Status: "running",
		//
		StartedAt: time.Now(),
	}

	// setup jobs
	for index, j := range s.Jobs {
		err := j.Setup(fmt.Sprintf("%s.%d", s.State.ID, index), &job.Job{
			Workdir: s.Workdir,
			//
			Image:       s.Image,
			Environment: s.Environment,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Run runs jobs in parallel
func (s *Stage) Run(ctx context.Context) error {
	s.logger.Infof("[stage: %s] start", s.Name)
	defer s.logger.Infof("[stage: %s] done", s.Name)

	g, ctx := errgroup.WithContext(ctx)

	for _, job := range s.Jobs {
		g.Go(func() error {
			return job.Run(ctx)
		})
	}

	if err := g.Wait(); err != nil {
		s.State.Status = "failed"
		s.State.Error = err.Error()
		s.State.FailedAt = time.Now()
		return err
	}

	s.State.Status = "succeeded"
	s.State.SucceedAt = time.Now()

	return nil
}

func (s *Stage) SetStdout(stdout io.Writer) {
	for _, job := range s.Jobs {
		job.SetStdout(stdout)
	}
}

func (s *Stage) SetStderr(stderr io.Writer) {
	for _, job := range s.Jobs {
		job.SetStderr(stderr)
	}
}
