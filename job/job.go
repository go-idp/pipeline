package job

import (
	"io"

	"github.com/go-idp/pipeline/step"
	"github.com/go-zoox/logger"
)

type Job struct {
	Name  string       `json:"name" yaml:"name"`
	Steps []*step.Step `json:"steps" yaml:"steps"`
	//
	Workdir string `json:"workdir" yaml:"workdir"`
	//
	Image string `json:"image" yaml:"image"`
	//
	ImageRegistry         string `json:"image_registry" yaml:"image_registry"`
	ImageRegistryUsername string `json:"image_registry_username" yaml:"image_registry_username"`
	ImageRegistryPassword string `json:"image_registry_password" yaml:"image_registry_password"`
	//
	Environment map[string]string `json:"environment" yaml:"environment"`
	//
	Timeout int64 `json:"timeout" yaml:"timeout"`
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
