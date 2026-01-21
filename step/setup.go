package step

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/go-zoox/core-utils/strings"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
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
		if s.Service.Name == "" {
			return fmt.Errorf("service name is required for docker-compose")
		}

		if s.Service.Version == "" {
			return fmt.Errorf("service version is required for docker-compose")
		}

		s.logger.Infof("[workflow][service] use service(type: %s, name: %s) in step(%s)", s.Service.Type, s.Service.Name, s.Name)

		s.Environment["PIPELINE_SERVICE_TYPE"] = s.Service.Type
		s.Environment["PIPELINE_SERVICE_NAME"] = s.Service.Name
		s.Environment["PIPELINE_SERVICE_VERSION"] = s.Service.Version
		// s.Environment["PIPELINE_SERVICE_CONFIG"] = base64.StdEncoding.EncodeToString([]byte(s.Service.Config))

		// v1 => use command
		if s.Service.Version == "v1" {
			commands := []string{}
			isTmp := false
			if ok := fs.IsExist(s.Service.Config); ok {
				s.Environment["PIPELINE_SERVICE_CONFIG_FILE"] = s.Service.Config
				commands = append(commands, fmt.Sprintf("export PIPELINE_SERVICE_CONFIG_FILE=%s", s.Service.Config))
			} else {
				// @TODO write config to tmp file
				commands = append(commands, fmt.Sprintf(`
PIPELINE_SERVICE_CONFIG_FILE=$(mktemp)

cat <<EOF > $PIPELINE_SERVICE_CONFIG_FILE
%s
EOF`, s.Service.Config))
				isTmp = true
			}

			switch s.Service.Type {
			case "docker-compose":

				commands = append(commands, fmt.Sprintf("export COMPOSE_PROJECT_NAME=\"%s\"", s.Service.Name))
				commands = append(commands, "docker-compose -f $PIPELINE_SERVICE_CONFIG_FILE up -d")
			case "docker-swarm":
				// commands = append(commands, fmt.Sprintf("export COMPOSE_PROJECT_NAME=\"%s\"", s.Service.Name))
				commands = append(commands, fmt.Sprintf("docker stack deploy --with-registry-auth -c $PIPELINE_SERVICE_CONFIG_FILE %s", s.Service.Name))
			case "kubernetes":
				commands = append(commands, "kubectl apply -f $PIPELINE_SERVICE_CONFIG_FILE")
			default:
				return fmt.Errorf("unsupported service type %s, only support docker-compose | docker-swarm | kubernetes", s.Service.Type)
			}

			if isTmp {
				// remove the tmp file
				commands = append(commands, `rm -f $PIPELINE_SERVICE_CONFIG_FILE`)
			}

			// generate the command
			s.Command = strings.Join(commands, "\n\n")

			logger.Debugf("[workflow][service] %s", s.Command)
		} else if s.Service.Version == "v2" {
			// @TODO v2 => use sdk
			return fmt.Errorf("service version %s is working in progress currently", s.Service.Version)
		} else {
			return fmt.Errorf("unsupported service version %s, only support v1 | v2", s.Service.Version)
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
