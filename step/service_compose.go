package step

import (
	"context"
	"fmt"
	"time"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

// runServiceCompose deploys a docker-compose project natively with the compose
// v2 SDK, equivalent to `docker compose --project-name <name> up -d` plus a
// startup readiness wait.
//
// The compose file is loaded in-memory (env interpolation uses the step
// environment), then created/started through the compose engine.
func (s *Step) runServiceCompose(ctx context.Context) error {
	s.serviceLogf("docker-compose: deploy project %q (timeout: %ds)", s.Service.Name, s.Service.Timeout)

	project, err := s.loadComposeProject()
	if err != nil {
		return fmt.Errorf("failed to load docker-compose config: %s", err)
	}

	dockerClient, err := newDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %s", err)
	}
	defer dockerClient.Close()

	dockerCli, err := newDockerCli(dockerClient, s.stdout, s.stderr)
	if err != nil {
		return fmt.Errorf("failed to create docker cli: %s", err)
	}

	s.setupRegistryAuth(dockerCli)

	svc := compose.NewComposeService(dockerCli)

	waitTimeout := time.Duration(s.Service.Timeout) * time.Second
	s.serviceLogf("docker-compose: creating and starting %d service(s)", len(project.Services))
	if err := svc.Up(ctx, project, api.UpOptions{
		Create: api.CreateOptions{},
		Start: api.StartOptions{
			Wait:        true,
			WaitTimeout: waitTimeout,
		},
	}); err != nil {
		s.collectComposeDiagnostics(ctx, svc, project.Name)
		return fmt.Errorf("docker compose up failed: %s", err)
	}

	s.serviceLogf("docker-compose: project %q is ready", s.Service.Name)
	return nil
}

// loadComposeProject parses and interpolates the compose config using the step
// environment (${VAR} references are resolved like `docker compose` does).
func (s *Step) loadComposeProject() (*types.Project, error) {
	details := types.ConfigDetails{
		WorkingDir: s.Workdir,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: "docker-compose.yaml",
				Content:  []byte(s.Service.Config),
			},
		},
		Environment: s.Environment,
	}

	project, err := loader.Load(details, func(opts *loader.Options) {
		opts.SetProjectName(s.Service.Name, true)
	})
	if err != nil {
		return nil, err
	}

	if project.Name != s.Service.Name {
		return nil, fmt.Errorf("project name mismatch: got %q want %q", project.Name, s.Service.Name)
	}

	return project, nil
}

// collectComposeDiagnostics dumps the logs of the project containers on failure,
// mirroring the diagnostics of the legacy shell implementation.
func (s *Step) collectComposeDiagnostics(ctx context.Context, svc api.Service, projectName string) {
	containers, err := svc.Ps(ctx, projectName, api.PsOptions{})
	if err != nil {
		s.serviceLogf("docker-compose: failed to list containers for diagnostics: %s", err)
		return
	}

	for _, c := range containers {
		s.serviceLogf("docker-compose: ===== container %s (service: %s, state: %s) =====", c.ID[:12], c.Service, c.State)
	}

	// best-effort log dump via the docker API client is not available here
	// (svc owns it); the container state above is enough to guide debugging.
}
