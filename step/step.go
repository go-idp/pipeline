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
	Plugin *Plugin `json:"plugin" yaml:"plugin"`
	//
	State *State `json:"state" yaml:"state"`
	//
	stdout io.Writer
	stderr io.Writer
	//
	logger *logger.Logger
}

// Plugin represents a plugin of the step
type Plugin struct {
	// Image is the image of the plugin, e.g. "docker.io/library/alpine:latest"
	Image string `json:"image" yaml:"image"`

	// Settings are the settings of the plugin
	// rules: PIPELINE_PLUGIN_SETTINGS_<snake case of key>=value
	// e.g. {"key": "value", "a": "b" } => PIPELINE_PLUGIN_SETTINGS_KEY=value, PIPELINE_PLUGIN_SETTINGS_A=b
	Settings map[string]string `json:"settings" yaml:"settings"`

	// Entrypoint is the entrypoint of the plugin, default is "/pipeline/plugin/run"
	Entrypoint string `json:"entrypoint" yaml:"entrypoint"`
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
