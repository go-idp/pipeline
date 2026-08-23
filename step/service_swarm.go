package step

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/cli/cli/command/stack/loader"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/command/stack/swarm"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/spf13/pflag"
)

// runServiceSwarm deploys a docker-compose definition as a swarm stack
// natively with the docker CLI stack deploy SDK, equivalent to
// `docker stack deploy --prune --with-registry-auth -c <config> <name>`.
//
// The deploy itself is detached (returns once accepted); readiness is then
// polled with the service timeout, mirroring the legacy shell behavior.
func (s *Step) runServiceSwarm(ctx context.Context) error {
	s.serviceLogf("docker-swarm: deploy stack %q (timeout: %ds)", s.Service.Name, s.Service.Timeout)

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

	// the stack deploy SDK reads compose files from disk
	tmpDir, err := os.MkdirTemp("", "pipeline-stack-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	composeFile := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composeFile, []byte(s.Service.Config), 0o644); err != nil {
		return fmt.Errorf("failed to write compose file: %s", err)
	}

	opts := options.Deploy{
		Composefiles:     []string{composeFile},
		Namespace:        s.Service.Name,
		ResolveImage:     swarm.ResolveImageAlways,
		SendRegistryAuth: true,
		Prune:            true,
		Detach:           true,
		Quiet:            true,
	}

	flags := pflag.NewFlagSet("stack deploy", pflag.ContinueOnError)
	flags.Bool("detach", true, "")
	flags.Bool("prune", true, "")
	flags.Bool("with-registry-auth", true, "")
	flags.Bool("quiet", true, "")
	flags.String("resolve-image", swarm.ResolveImageAlways, "")
	// mark detach as explicitly set to suppress the CLI's detach warning
	_ = flags.Set("detach", "true")

	cfg, err := loader.LoadComposefile(dockerCli, opts)
	if err != nil {
		return fmt.Errorf("failed to load compose config for stack deploy: %s", err)
	}

	if err := swarm.RunDeploy(ctx, dockerCli, flags, &opts, cfg); err != nil {
		s.collectSwarmDiagnostics(ctx, dockerClient)
		return fmt.Errorf("docker stack deploy failed: %s", err)
	}

	if err := s.waitSwarmReady(ctx, dockerClient); err != nil {
		s.collectSwarmDiagnostics(ctx, dockerClient)
		return err
	}

	s.serviceLogf("docker-swarm: stack %q is ready", s.Service.Name)
	return nil
}

// stackFilter returns the docker filters selecting resources of this stack.
func (s *Step) stackFilter() filters.Args {
	return filters.NewArgs(filters.Arg("label", fmt.Sprintf("com.docker.stack.namespace=%s", s.Service.Name)))
}

// waitSwarmReady polls the stack services until every service has its desired
// replicas running, bounded by the service timeout.
func (s *Step) waitSwarmReady(ctx context.Context, dockerClient *client.Client) error {
	timeout := time.Duration(s.Service.Timeout) * time.Second
	interval := 6 * time.Second

	deadline := time.Now().Add(timeout)
	checks := 0
	for {
		checks++
		ready, notReady, err := s.swarmReadyState(ctx, dockerClient)
		if err != nil {
			return fmt.Errorf("failed to check swarm deployment status: %s", err)
		}

		if ready {
			s.serviceLogf("docker-swarm: deployment ready (check %d)", checks)
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("swarm deployment startup timeout after %d checks, not ready: %s", checks, strings.Join(notReady, "; "))
		}

		s.serviceLogf("docker-swarm: waiting for deployment startup (check %d/%d): %s", checks, int(timeout/interval), strings.Join(notReady, "; "))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

// swarmReadyState returns whether all stack services have their desired
// replicas running, plus a description of services that are not ready yet.
func (s *Step) swarmReadyState(ctx context.Context, dockerClient *client.Client) (bool, []string, error) {
	services, err := dockerClient.ServiceList(ctx, types.ServiceListOptions{Filters: s.stackFilter()})
	if err != nil {
		return false, nil, err
	}

	ready := true
	var notReady []string

	for _, svc := range services {
		name := strings.TrimPrefix(svc.Spec.Name, s.Service.Name+"_")
		var desired uint64
		var running uint64

		if svc.Spec.Mode.Replicated != nil && svc.Spec.Mode.Replicated.Replicas != nil {
			desired = *svc.Spec.Mode.Replicated.Replicas
		} else if svc.Spec.Mode.Global != nil {
			// global mode: at least one running task per node
			desired = 1
		}

		if svc.ServiceStatus != nil {
			running = svc.ServiceStatus.RunningTasks
		}

		if desired == 0 || running < desired {
			ready = false
			notReady = append(notReady, fmt.Sprintf("%s (%d/%d)", name, running, desired))
		}
	}

	return ready, notReady, nil
}

// collectSwarmDiagnostics prints the stack services and failed tasks, mirroring
// the diagnostics of the legacy shell implementation.
func (s *Step) collectSwarmDiagnostics(ctx context.Context, dockerClient *client.Client) {
	s.serviceLogf("docker-swarm: collecting stack diagnostics ...")

	services, err := dockerClient.ServiceList(ctx, types.ServiceListOptions{Filters: s.stackFilter()})
	if err != nil {
		s.serviceLogf("docker-swarm: failed to list stack services: %s", err)
		return
	}

	for _, svc := range services {
		name := strings.TrimPrefix(svc.Spec.Name, s.Service.Name+"_")
		s.serviceLogf("docker-swarm: ==== service %s (image: %s) ====", name, svc.Spec.TaskTemplate.ContainerSpec.Image)

		tasks, err := dockerClient.TaskList(ctx, types.TaskListOptions{Filters: filters.NewArgs(filters.Arg("service", svc.ID))})
		if err != nil {
			s.serviceLogf("docker-swarm: failed to list tasks of %s: %s", name, err)
			continue
		}

		for _, t := range tasks {
			s.serviceLogf("docker-swarm: task %s state=%s err=%q image=%s", t.ID[:12], t.Status.State, t.Status.Err, t.Spec.ContainerSpec.Image)

			if t.Status.State == swarmtypes.TaskStateFailed || t.Status.State == swarmtypes.TaskStateRejected {
				s.dumpServiceLogs(ctx, dockerClient, svc.ID, name)
			}
		}
	}
}

// dumpServiceLogs prints the tail of a swarm service's logs (best effort).
func (s *Step) dumpServiceLogs(ctx context.Context, dockerClient *client.Client, serviceID, name string) {
	reader, err := dockerClient.ServiceLogs(ctx, serviceID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "100",
		Details:    false,
	})
	if err != nil {
		s.serviceLogf("docker-swarm: failed to get logs of service %s: %s", name, err)
		return
	}
	defer reader.Close()

	s.serviceLogf("docker-swarm: ---- logs of service %s (tail 100) ----", name)
	if _, err := io.Copy(s.stdout, reader); err != nil {
		s.serviceLogf("docker-swarm: failed to stream logs of service %s: %s", name, err)
	}
}
