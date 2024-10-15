package stage

import (
	"io"

	"github.com/go-idp/pipeline/job"
	"github.com/go-zoox/logger"
)

type Stage struct {
	Name string     `json:"name" yaml:"name"`
	Jobs []*job.Job `json:"jobs" yaml:"jobs"`
	//
	Workdir string `json:"workdir" yaml:"workdir"`
	//
	Image       string            `json:"image" yaml:"image"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	// JobRunMode is the mode of the job, e.g. "serial", "parallel", default: parallel
	JobRunMode string `json:"job_run_mode" yaml:"job_run_mode"`
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
