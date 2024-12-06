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
	Engine string `json:"engine" yaml:"engine"`
	//
	Image string `json:"image" yaml:"image"`
	//
	Shell string `json:"shell" yaml:"shell"`
	// Timeout is the timeout of the step, unit: second, default: 86400 (1 day)
	Timeout int64 `json:"timeout" yaml:"timeout"`
	//
	Plugin *Plugin `json:"plugin" yaml:"plugin"`
	//
	Language *Language `json:"language" yaml:"language"`
	//
	Service *Service `json:"service" yaml:"service"`
	//
	State *State `json:"state" yaml:"state"`
	//
	stdout io.Writer
	stderr io.Writer
	//
	logger *logger.Logger
}

// Language represents a language of the step
type Language struct {
	// Name is the name of the language, e.g. "node", "go", "python"
	Name string `json:"name" yaml:"name"`

	// Version is the version of the language, e.g. "12", "1.20", "3.10"
	Version string `json:"version" yaml:"version"`
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
