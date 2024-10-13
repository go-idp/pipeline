package step

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-zoox/command"
	"github.com/go-zoox/command/config"
)

// RunConfig is the config for run
type RunConfig struct {
	// Total is the total count of the parent steps
	Total int
	// Current is the current index of the parent steps
	Current int
	// Parent is the parent name
	Parent string
}

// RunOption is the option for run
type RunOption func(cfg *RunConfig)

// Run runs the step
func (s *Step) Run(ctx context.Context, opts ...RunOption) error {
	cfg := &RunConfig{}
	for _, o := range opts {
		o(cfg)
	}

	s.logger.Infof("%s[step(%d/%d): %s] start", cfg.Parent, cfg.Current, cfg.Total, s.Name)
	defer s.logger.Infof("%s[step(%d/%d): %s] done", cfg.Parent, cfg.Current, cfg.Total, s.Name)

	if s.Plugin != nil {
		s.logger.Infof("%s[step(%d/%d): %s] use plugin => %s", cfg.Parent, cfg.Current, cfg.Total, s.Name, s.Plugin.Image)
	}

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
