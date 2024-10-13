package step

import (
	"fmt"
	"os"
	"time"

	"github.com/go-zoox/core-utils/strings"
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

	if s.Plugin != nil {
		// s.logger.Infof("[workflow][plugin] use %s in step(%s)", s.Plugin.Image, s.Name)

		if s.Plugin.Entrypoint == "" {
			s.Plugin.Entrypoint = "/pipeline/plugin/run"
		}

		s.Image = s.Plugin.Image

		// Check if /pipeline/plugin/run exists, if not, return an error
		s.Command = fmt.Sprintf(`if [ ! -f "%s" ]; then echo -e "\033[0;31merror: it is not a pipeline plugin (%s not found)\033[0m"; exit 127; fi`, s.Plugin.Entrypoint, s.Plugin.Entrypoint)

		// Settings are passed as environment variables
		// e.g. {"key": "value" } => -e PIPELINE_PLUGIN_SETTINGS_KEY=value
		for k, v := range s.Plugin.Settings {
			s.Environment["PIPELINE_PLUGIN_SETTINGS_"+strings.UpperCase(k)] = v
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
