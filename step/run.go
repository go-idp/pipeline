package step

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-zoox/command"
	"github.com/go-zoox/command/config"
	"github.com/go-zoox/core-utils/strings"
	"github.com/go-zoox/crypto/base64"
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

	ccfg := &config.Config{
		Context: ctx,
		//
		Command:     s.Command,
		Environment: s.Environment,
		//
		WorkDir: s.Workdir,
		//
		Image: s.Image,
		//
		Shell: s.Shell,
	}

	ccfg.Engine = "host"
	if s.Image != "" {
		ccfg.Engine = "docker"
	}

	if s.Engine != "" {
		ccfg.Engine = s.Engine

		//
		agentX, err := url.Parse(s.Engine)
		if err == nil {
			ccfg.Engine = agentX.Scheme

			ccfg.Server = agentX.Host
			ccfg.ClientID = agentX.User.Username()
			ccfg.ClientSecret, _ = agentX.User.Password()

			// @TODO
			switch ccfg.Engine {
			case "host":
			case "docker":
			case "ssh":
				ccfg.SSHHost = agentX.Hostname()
				ccfg.SSHPort = strings.MustToInt(agentX.Port())
				ccfg.SSHUser = ccfg.ClientID
				ccfg.SSHPass = ccfg.ClientSecret
				ccfg.SSHIsIgnoreStrictHostKeyChecking = true

				if ccfg.SSHUser == "private_key" {
					if ccfg.SSHPass == "" {
						return fmt.Errorf("private_key should be set for ssh engine, when user is private_key")
					}

					ccfg.SSHPrivateKey = base64.Decode(ccfg.SSHPass)
				}
			case "idp":
				ccfg.Server = fmt.Sprintf("ws://%s", agentX.Host)
			case "idps":
				ccfg.Server = fmt.Sprintf("wss://%s", agentX.Host)
				ccfg.Engine = "idp"
			default:
				return fmt.Errorf("unsupported engine: %s (uri: %s)", ccfg.Engine, s.Engine)
			}
		}
	}

	cmd, err := command.New(ccfg)
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
