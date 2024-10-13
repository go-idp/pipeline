package step

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/go-zoox/command"
	"github.com/go-zoox/command/config"
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

type RunConfig struct {
	Total   int
	Current int
	//
	Parent string
}

type RunOption func(cfg *RunConfig)

func (s *Step) getLogger() *logger.Logger {
	l := logger.New()
	l.SetStdout(s.stdout)
	return l
}

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

func (s *Step) Run(ctx context.Context, opts ...RunOption) error {
	cfg := &RunConfig{}
	for _, o := range opts {
		o(cfg)
	}

	s.logger.Infof("%s[step(%d/%d): %s] start", cfg.Parent, cfg.Current, cfg.Total, s.Name)
	defer s.logger.Infof("%s[step(%d/%d): %s] done", cfg.Parent, cfg.Current, cfg.Total, s.Name)

	if s.State == nil {
		return fmt.Errorf("you should setup before run")
	}

	engine := "host"
	if s.Image != "" {
		engine = "docker"
	}

	agentServer := "ws://192.168.31.246:8838"
	agentUsername := ""
	agentPassword := ""
	if s.Agent != "" {
		agentX, err := url.Parse(s.Agent)
		if err != nil {
			return fmt.Errorf("failed to parse agent: %s", err)
		}

		engine = agentX.Scheme
		// agentServer = agentX.Host
		agentUsername = agentX.User.Username()
		agentPassword, _ = agentX.User.Password()
	}

	cmd, err := command.New(&config.Config{
		Context: ctx,
		//
		Command:     s.Command,
		Environment: s.Environment,
		//
		WorkDir: s.Workdir,
		//
		Image:  s.Image,
		Engine: engine,
		//
		Shell: s.Shell,
		// IsInheritEnvironmentEnabled: true,
		Server:       agentServer,
		ClientID:     agentUsername,
		ClientSecret: agentPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to create command: %s", err)
	}

	if err := cmd.SetStdout(s.stdout); err != nil {
		return fmt.Errorf("failed to set stdout: %s", err)
	}

	if err := cmd.SetStderr(s.stderr); err != nil {
		return fmt.Errorf("failed to set stderr: %s", err)
	}

	if err := cmd.Run(); err != nil {
		s.State.Status = "failed"
		s.State.Error = err.Error()
		s.State.FailedAt = time.Now()
		// s.State.ExitCode = cmd.Cancel()
		return fmt.Errorf("failed to run command: %s", err)
	}
	s.State.Status = "succeeded"
	s.State.SucceedAt = time.Now()

	return nil
}

func (s *Step) SetStdout(stdout io.Writer) {
	s.stdout = stdout
}

func (s *Step) SetStderr(stderr io.Writer) {
	s.stderr = stderr
}
