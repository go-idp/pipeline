package step

import (
	"os"
	"time"
)

// Setup sets up the step
func (s *Step) Setup(id string, opts ...*Step) error {
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
		ID:        id,
		Status:    "running",
		StartedAt: time.Now(),
	}

	return nil
}
