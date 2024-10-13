package stage

import (
	"fmt"
	"os"
	"time"

	"github.com/go-idp/pipeline/job"
)

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
			Environment: s.Environment,
			//
			Image: s.Image,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
