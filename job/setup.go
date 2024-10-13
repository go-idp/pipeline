package job

import (
	"fmt"
	"os"
	"time"

	"github.com/go-idp/pipeline/step"
)

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
