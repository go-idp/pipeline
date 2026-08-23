package step

import (
	"encoding/base64"
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

		if s.ImageRegistry == "" {
			s.ImageRegistry = opt.ImageRegistry
		}

		if s.ImageRegistryUsername == "" {
			s.ImageRegistryUsername = opt.ImageRegistryUsername
		}

		if s.ImageRegistryPassword == "" {
			s.ImageRegistryPassword = opt.ImageRegistryPassword
		}

		if s.Workdir == "" {
			s.Workdir = opt.Workdir
		}

		if s.Shell == "" {
			s.Shell = opt.Shell
		}

		if s.Timeout == 0 {
			s.Timeout = opt.Timeout
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

		if s.DataDirInner == "" {
			s.DataDirInner = opt.DataDirInner
		}

		if s.DataDirOuter == "" {
			s.DataDirOuter = opt.DataDirOuter
		}

		if s.Observe == nil {
			s.Observe = opt.Observe
		}
	}

	// environment
	if s.Environment == nil {
		s.Environment = map[string]string{}
	}

	// default timeout is 1 day
	if s.Timeout == 0 {
		s.Timeout = 86400
	}

	// if language is set, will use the language
	if s.Language != nil {
		if s.Plugin != nil {
			return fmt.Errorf("you can not use language and plugin at the same time")
		}

		s.logger.Infof("[workflow][language] use %s in step(%s)", s.Language.Name, s.Name)
		s.Plugin = &Plugin{
			Image: fmt.Sprintf("ghcr.io/go-idp/pipeline-language-%s:%s", s.Language.Name, s.Language.Version),
			// inherit the environment of the step
			inheritEnv: true,
		}
	}

	if s.Plugin != nil {
		//
		originCommand := s.Command
		originEnvironment := s.Environment

		// s.logger.Infof("[workflow][plugin] use %s in step(%s)", s.Plugin.Image, s.Name)

		if s.Plugin.Entrypoint == "" {
			s.Plugin.Entrypoint = "/pipeline/plugin/run"
		}

		s.Image = s.Plugin.Image

		if s.ImageRegistry != "" {
			s.ImageRegistry = s.Plugin.ImageRegistry
		}

		if s.ImageRegistryUsername != "" {
			s.ImageRegistryUsername = s.Plugin.ImageRegistryUsername
		}

		if s.ImageRegistryPassword != "" {
			s.ImageRegistryPassword = s.Plugin.ImageRegistryPassword
		}

		// Check if /pipeline/plugin/run exists, if not, return an error
		s.Command = fmt.Sprintf(
			`if [ ! -f "%s" ]; then echo -e "\033[0;31merror: it is not a pipeline plugin (%s not found)\033[0m"; exit 127; fi; %s`,
			s.Plugin.Entrypoint,
			s.Plugin.Entrypoint,
			s.Plugin.Entrypoint,
		)

		// Settings are passed as environment variables
		// will reset the environment
		// e.g. {"key": "value" } => -e PIPELINE_PLUGIN_SETTINGS_KEY=value
		//  value support environment variables, e.g. {"key": "${ENV}" } => -e PIPELINE_PLUGIN_SETTINGS_KEY=${ENV}
		s.Environment = map[string]string{}
		// interhit the environment
		if s.Plugin.inheritEnv {
			s.Environment = originEnvironment
		}
		//
		s.Environment["PIPELINE_PLUGIN_COMMAND"] = base64.StdEncoding.EncodeToString([]byte(originCommand))
		//
		for k, v := range s.Plugin.Settings {
			// if value is environment variable, replace it
			if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
				key := strings.TrimPrefix(strings.TrimSuffix(v, "}"), "${")
				if val, ok := originEnvironment[key]; ok {
					s.Environment["PIPELINE_PLUGIN_SETTINGS_"+strings.UpperCase(k)] = val
				}
			} else {
				s.Environment["PIPELINE_PLUGIN_SETTINGS_"+strings.UpperCase(k)] = v
			}
		}
	}

	if s.Service != nil {
		if s.Service.Type == "" {
			return fmt.Errorf("service type is required, only support docker-compose | docker-swarm | kubernetes")
		}

		if s.Service.Name == "" {
			return fmt.Errorf("service name is required")
		}

		if s.Service.Config == "" {
			return fmt.Errorf("service config is required")
		}

		switch s.Service.Type {
		case "docker-compose", "docker-swarm", "kubernetes", "k8s":
		default:
			return fmt.Errorf("unsupported service type %s, only support docker-compose | docker-swarm | kubernetes", s.Service.Type)
		}

		if s.Service.Timeout == 0 {
			s.Service.Timeout = 120
		}

		// interpolate ${VAR} references in service scalar fields with the
		// merged step environment (same as the `config` interpolation)
		s.Service.Name = expandStepEnv(s.Service.Name, s.Environment)
		s.Service.ImageRegistry = expandStepEnv(s.Service.ImageRegistry, s.Environment)
		s.Service.ImageRegistryUsername = expandStepEnv(s.Service.ImageRegistryUsername, s.Environment)
		s.Service.ImageRegistryPassword = expandStepEnv(s.Service.ImageRegistryPassword, s.Environment)
		s.Service.Namespace = expandStepEnv(s.Service.Namespace, s.Environment)
		s.Service.Kubeconfig = expandStepEnv(s.Service.Kubeconfig, s.Environment)

		s.logger.Infof("[workflow][service] use service(type: %s, name: %s) in step(%s)", s.Service.Type, s.Service.Name, s.Name)
	}

	// setup state
	s.State = &State{
		ID:        id,
		Status:    "pending",
		StartedAt: time.Now(),
	}

	return nil
}

// expandStepEnv replaces ${KEY} references with values from the environment.
// Missing keys are left as-is (surfaced later by validation/execution).
func expandStepEnv(value string, env map[string]string) string {
	if !strings.Contains(value, "${") {
		return value
	}

	for k, v := range env {
		value = strings.ReplaceAll(value, "${"+k+"}", v)
	}

	return value
}
