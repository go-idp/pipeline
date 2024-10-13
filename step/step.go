package step

import (
	"io"

	"github.com/go-zoox/logger"
)

type Step struct {
	Name string `json:"name" yaml:"name"`
	//
	Command     string            `json:"command" yaml:"command"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	//
	Workdir string `json:"workdir" yaml:"workdir"`
	//
	Agent string `json:"agent" yaml:"agent"`
	//
	Image string `json:"image" yaml:"image"`
	//
	Shell string `json:"shell" yaml:"shell"`
	//
	State *State `json:"state" yaml:"state"`
	//
	stdout io.Writer
	stderr io.Writer
	//
	logger *logger.Logger
}

func (s *Step) getLogger() *logger.Logger {
	l := logger.New()
	l.SetStdout(s.stdout)
	return l
}

// SetStdout sets the stdout of the step
func (s *Step) SetStdout(stdout io.Writer) {
	s.stdout = stdout
}

// SetStderr sets the stderr of the step
func (s *Step) SetStderr(stderr io.Writer) {
	s.stderr = stderr
}
