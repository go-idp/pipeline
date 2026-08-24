package step

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"
	cliconfigtypes "github.com/docker/cli/cli/config/types"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
)

// Service types
const (
	ServiceTypeDockerCompose = "docker-compose"
	ServiceTypeDockerSwarm   = "docker-swarm"
	ServiceTypeKubernetes    = "kubernetes"
)

// runService deploys the step service natively with Go SDKs.
// The SDK calls run in the pipeline process (the agent host that executes
// `pipeline run`), which is the same host that has docker / kubectl access.
func (s *Step) runService(ctx context.Context) error {
	switch s.Service.Type {
	case ServiceTypeDockerCompose:
		return s.runServiceCompose(ctx)
	case ServiceTypeDockerSwarm:
		return s.runServiceSwarm(ctx)
	case ServiceTypeKubernetes, "k8s":
		return s.runServiceKubernetes(ctx)
	default:
		return fmt.Errorf("unsupported service type %s", s.Service.Type)
	}
}

// serviceLogf writes a log line to the step stdout (with timestamp).
func (s *Step) serviceLogf(format string, args ...any) {
	if s.stdout == nil {
		return
	}
	fmt.Fprintf(s.stdout, "[%s][service] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
}

// newDockerClient creates a docker API client from the environment
// (DOCKER_HOST / DOCKER_TLS_VERIFY / ~/.docker), with API version negotiation.
func newDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// newDockerCli builds a docker CLI object backed by the API client, which is
// required by both the compose SDK and the swarm stack deploy SDK.
func newDockerCli(dockerClient *client.Client, stdout, stderr io.Writer) (command.Cli, error) {
	dockerCli, err := command.NewDockerCli(
		command.WithAPIClient(dockerClient),
		command.WithOutputStream(stdout),
		command.WithErrorStream(stderr),
	)
	if err != nil {
		return nil, err
	}
	// Initialize is required: without it the CLI context is left unset
	// (currentContext == "" and contextStore == nil), so any call to
	// dockerCli.Client() fails with "unable to resolve docker endpoint: no
	// context store initialized" and os.Exit(1)s the whole process.
	if err := dockerCli.Initialize(&cliflags.ClientOptions{}); err != nil {
		return nil, err
	}
	return dockerCli, nil
}

// setupRegistryAuth injects the service registry credentials into the docker
// CLI config file, equivalent to `docker login <registry>`.
// compose pulls and swarm --with-registry-auth read auth from the config file.
func (s *Step) setupRegistryAuth(dockerCli command.Cli) {
	if s.Service.ImageRegistry == "" || s.Service.ImageRegistryUsername == "" || s.Service.ImageRegistryPassword == "" {
		return
	}

	auth := cliconfigtypes.AuthConfig{
		Username:      s.Service.ImageRegistryUsername,
		Password:      s.Service.ImageRegistryPassword,
		ServerAddress: s.Service.ImageRegistry,
	}

	cf := dockerCli.ConfigFile()
	if cf.AuthConfigs == nil {
		cf.AuthConfigs = map[string]cliconfigtypes.AuthConfig{}
	}

	// key candidates: raw value, scheme-stripped host, and the docker hub index key
	for _, key := range registryAuthKeys(s.Service.ImageRegistry) {
		cf.AuthConfigs[key] = auth
	}

	s.serviceLogf("registry auth configured for %s", normalizeRegistryHost(s.Service.ImageRegistry))
}

// registryAuthKeys returns the config-file keys to write the auth under.
func registryAuthKeys(registry string) []string {
	host := normalizeRegistryHost(registry)
	keys := []string{registry, host}

	switch host {
	case "docker.io", "index.docker.io", "registry-1.docker.io":
		keys = append(keys, "https://index.docker.io/v1/")
	}

	return keys
}

// normalizeRegistryHost strips scheme and trailing slashes from a registry
// address, e.g. "https://registry.example.com:5000/" => "registry.example.com:5000".
func normalizeRegistryHost(registry string) string {
	host := strings.TrimSpace(registry)
	if idx := strings.Index(host, "://"); idx >= 0 {
		host = host[idx+3:]
	}
	host = strings.TrimRight(host, "/")
	return host
}
