package step

import (
	"strings"
	"testing"
)

func TestServiceUsesCommand(t *testing.T) {
	cases := []struct {
		name   string
		step   *Step
		expect bool
	}{
		{name: "nil service", step: &Step{Engine: "ssh://u:p@h:22"}, expect: false},
		{name: "empty engine", step: &Step{Engine: "", Service: &Service{}}, expect: false},
		{name: "host engine", step: &Step{Engine: "host", Service: &Service{}}, expect: false},
		{name: "host literal", step: &Step{Engine: "host", Service: &Service{}}, expect: false},
		{name: "ssh engine", step: &Step{Engine: "ssh://user:pass@10.0.0.2:22", Service: &Service{}}, expect: true},
		{name: "idp engine", step: &Step{Engine: "idp://user:pass@10.0.0.2:8838", Service: &Service{}}, expect: true},
		{name: "docker engine", step: &Step{Engine: "docker://host", Service: &Service{}}, expect: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.step.serviceUsesCommand(); got != c.expect {
				t.Fatalf("serviceUsesCommand(%q) = %v, want %v", c.step.Engine, got, c.expect)
			}
		})
	}
}

func TestBuildServiceCommandCompose(t *testing.T) {
	s := &Step{Service: &Service{
		Type:                  "docker-compose",
		Name:                  "my-task-10086",
		Config:                "services:\n  web:\n    image: nginx:alpine",
		Timeout:               120,
		ImageRegistry:         "registry.example.com",
		ImageRegistryUsername: "deploy",
		ImageRegistryPassword: "secret",
	}}

	cmd, err := s.buildServiceCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"docker compose -f",
		"-p 'my-task-10086'",
		"up -d --wait --wait-timeout 120",
		"docker login",
		"--password-stdin",
		"PIPELINE_SERVICE_REGISTRY_PASS",
		"PIPELINE_SERVICE_EOF",
		"set -e",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("compose command missing %q\ncommand:\n%s", want, cmd)
		}
	}

	if strings.Contains(cmd, "secret") {
		t.Errorf("registry password leaked into command:\n%s", cmd)
	}
}

func TestBuildServiceCommandSwarm(t *testing.T) {
	s := &Step{Service: &Service{
		Type:    "docker-swarm",
		Name:    "my-stack",
		Config:  "services:\n  web:\n    image: nginx:alpine\n    deploy:\n      replicas: 2",
		Timeout: 150,
	}}

	cmd, err := s.buildServiceCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"docker stack deploy --detach=false --prune --with-registry-auth -c",
		"docker stack deploy --prune --with-registry-auth -c",
		"'my-stack'",
		"docker stack deploy --help",
		"grep -q -- '--detach'",
		"PIPELINE_STACK='my-stack'",
		"com.docker.stack.namespace=$PIPELINE_STACK",
		"date +%s",
		"PIPELINE_DEADLINE",
		"docker service ls",
		"RUNNING=${REPLICAS%%/*}",
		"sleep 6",
		"PIPELINE_SERVICE_EOF",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("swarm command missing %q\ncommand:\n%s", want, cmd)
		}
	}
}

// TestSwarmDeployDetachProbe guards the regression where the generated swarm
// command chose `docker stack deploy --detach=false` by parsing a hard-coded
// docker version (>= 17.05). `--detach` for `docker stack deploy` was only added
// in Docker Engine 26.0.0, so that probe wrongly used `--detach=false` on
// docker < 26.0 and failed with "unknown flag: --detach" (exit 125). The command
// must probe `--detach` support via `docker stack deploy --help` instead of a
// version string.
func TestSwarmDeployDetachProbe(t *testing.T) {
	s := &Step{Service: &Service{
		Type:    "docker-swarm",
		Name:    "stack-x",
		Config:  "services:\n  web:\n    image: nginx:alpine",
		Timeout: 90,
	}}

	cmd, err := s.buildServiceCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"docker stack deploy --help",
		"grep -q -- '--detach'",
		"docker stack deploy --detach=false --prune --with-registry-auth -c",
		"docker stack deploy --prune --with-registry-auth -c",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("swarm command missing %q\ncommand:\n%s", want, cmd)
		}
	}

	// The old version-string probe must be gone: `--detach` for `docker stack
	// deploy` only is only supported from Docker Engine 26.0.0.
	for _, old := range []string{"docker version --format", "17.05", "PIPELINE_SWARM_VERSION"} {
		if strings.Contains(cmd, old) {
			t.Errorf("swarm command still uses version-string probe %q\ncommand:\n%s", old, cmd)
		}
	}
}

func TestBuildServiceCommandKube(t *testing.T) {
	s := &Step{Service: &Service{
		Type:       "kubernetes",
		Config:     "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: web",
		Namespace:  "demo",
		Kubeconfig: "/tmp/kubeconfig",
		Timeout:    90,
	}}

	cmd, err := s.buildServiceCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		`export KUBECONFIG="/tmp/kubeconfig"`,
		"kubectl apply -f",
		"-n 'demo'",
		"--for=condition=Available deployment --all",
		"--timeout=90s",
		"PIPELINE_SERVICE_EOF",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("kube command missing %q\ncommand:\n%s", want, cmd)
		}
	}
}

func TestSetup_RemoteServiceSetsCommand(t *testing.T) {
	s := &Step{
		Name:   "deploy",
		Engine: "ssh://user:pass@10.0.0.2:22",
		Service: &Service{
			Type:                  "docker-compose",
			Name:                  "stack-x",
			Config:                "services:\n  web:\n    image: nginx:alpine",
			ImageRegistry:         "registry.example.com",
			ImageRegistryUsername: "deploy",
			ImageRegistryPassword: "s3cret",
		},
	}

	if err := s.Setup("0"); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	if s.Command == "" {
		t.Fatal("remote service should populate s.Command")
	}
	if !strings.Contains(s.Command, "docker compose -f") {
		t.Errorf("expected docker compose command, got:\n%s", s.Command)
	}
	// registry credentials are forwarded via env, never inside the command
	if _, ok := s.Environment["PIPELINE_SERVICE_REGISTRY_PASS"]; !ok {
		t.Error("expected registry password env to be set")
	}
	if strings.Contains(s.Command, "s3cret") {
		t.Errorf("registry password leaked into command:\n%s", s.Command)
	}
}

func TestSetup_LocalServiceStaysSdk(t *testing.T) {
	s := &Step{
		Name: "deploy",
		Service: &Service{
			Type:   "docker-compose",
			Name:   "stack-x",
			Config: "services: {}",
		},
	}

	if err := s.Setup("0"); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	// a local (engine empty / host) service runs the Go SDK, so no command is set
	if s.Command != "" {
		t.Errorf("local service should not set a command (SDK path), got %q", s.Command)
	}
}

func TestEngineVersionAtLeast(t *testing.T) {
	cases := []struct {
		version      string
		major, minor int
		expect       bool
	}{
		{"27.3.1", 26, 0, true},
		{"26.0.0", 26, 0, true},
		{"26.1.0", 26, 0, true},
		{"25.0.6", 26, 0, false},
		{"24.0.7", 26, 0, false},
		{"20.10.22", 26, 0, false},
		{"17.05.0-ce", 26, 0, false},
		{"1.13.1", 26, 0, false},
		{"", 26, 0, false},
		{"abc", 26, 0, false},
	}
	for _, c := range cases {
		if got := engineVersionAtLeast(c.version, c.major, c.minor); got != c.expect {
			t.Errorf("engineVersionAtLeast(%q, %d, %d) = %v, want %v", c.version, c.major, c.minor, got, c.expect)
		}
	}
}

// TestServiceCommand_PathBootstrap guards the regression where the generated
// remote command ran `docker` / `kubectl` before ensuring the binary is on PATH.
// A non-login /bin/sh (used by the remote engine/agent) does not include the
// Homebrew bin dirs on macOS, so docker installed at /opt/homebrew/bin (Apple
// Silicon) or /usr/local/bin (Intel) is "command not found" (exit 127). The
// command must prepend those dirs to PATH when the binary is missing.
func TestServiceCommand_PathBootstrap(t *testing.T) {
	t.Run("docker (compose/swarm)", func(t *testing.T) {
		s := &Step{Service: &Service{
			Type:    "docker-swarm",
			Name:    "my-stack",
			Config:  "services:\n  web:\n    image: nginx:alpine",
			Timeout: 60,
		}}
		cmd, err := s.buildServiceCommand()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, want := range []string{
			"command -v docker",
			"/opt/homebrew/bin",
			"/usr/local/bin",
			`export PATH="$_bin:$PATH"`,
		} {
			if !strings.Contains(cmd, want) {
				t.Errorf("command missing %q\ncommand:\n%s", want, cmd)
			}
		}
	})

	t.Run("kubectl (k8s)", func(t *testing.T) {
		s := &Step{Service: &Service{
			Type:    "kubernetes",
			Config:  "apiVersion: v1\nkind: Service\nmetadata:\n  name: web",
			Timeout: 60,
		}}
		cmd, err := s.buildServiceCommand()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(cmd, "command -v kubectl") {
			t.Errorf("kube command missing kubectl path bootstrap\ncommand:\n%s", cmd)
		}
	})
}

// TestServiceCommand_DockerHostDiscovery guards the regression where the
// generated remote command ran `docker` without pointing DOCKER_HOST at the
// daemon socket. On macOS the daemon socket is not at /var/run/docker.sock
// (Docker Desktop / OrbStack / Colima put it under the user's home), so the CLI
// fell back to the default context and failed with
// "failed to connect to the docker API at unix:///var/run/docker.sock ... no
// such file or directory". The command must probe the common socket paths (and
// the macOS console user's home / all /Users/* homes as a fallback), export
// DOCKER_HOST, and when the current HOME is not writable (e.g. /root) switch
// HOME to the docker user's home derived from the socket path so docker login /
// Compose plugin discovery work under one home.
func TestServiceCommand_DockerHostDiscovery(t *testing.T) {
	types := []string{"docker-compose", "docker-swarm"}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			s := &Step{Service: &Service{
				Type:    typ,
				Name:    "my-task",
				Config:  "services:\n  web:\n    image: nginx:alpine",
				Timeout: 60,
			}}

			cmd, err := s.buildServiceCommand()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range []string{
				`if [ -z "$DOCKER_HOST" ]`,
				`"$_home/.docker/run/docker.sock"`,
				`"$_home/.orbstack/run/docker.sock"`,
				`"$_home/.colima/default/docker.sock"`,
				`"$_home/.colima"/*/docker.sock`,
				`stat -f '%Su' /dev/console`,
				`export DOCKER_HOST="unix://$_sock"`,
				`[ -S /var/run/docker.sock ]`,
				`/Users/*`,
				`[ ! -w "${HOME:-/}" ]`,
				`export HOME="/Users/${_u%%/*}"`,
			} {
				if !strings.Contains(cmd, want) {
					t.Errorf("command missing %q\ncommand:\n%s", want, cmd)
				}
			}
		})
	}
}

func TestBuildServiceCommandDefaultNamespace(t *testing.T) {
	s := &Step{Service: &Service{
		Type:    "k8s",
		Config:  "apiVersion: v1\nkind: Service\nmetadata:\n  name: web",
		Timeout: 60,
	}}

	cmd, err := s.buildServiceCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(cmd, "-n 'default'") {
		t.Errorf("expected default namespace in command:\n%s", cmd)
	}
}

// TestServiceCommand_RegistryLoginRedirectsDockerConfig guards the regression
// where the generated remote deploy command ran `docker login` without first
// redirecting DOCKER_CONFIG. On hosts whose HOME is read-only (e.g. macOS agents
// whose HOME is /root), `docker login` failed with "error saving credentials:
// mkdir /root: read-only file system". The command must redirect DOCKER_CONFIG to
// a writable temp dir (so login succeeds regardless of HOME) and symlink the
// user-level CLI plugins (~/.docker/cli-plugins, e.g. the Colima Compose plugin)
// into it, so `docker compose -f ...` stays discoverable (otherwise it fails with
// "unknown shorthand flag: 'f' in -f").
func TestServiceCommand_RegistryLoginRedirectsDockerConfig(t *testing.T) {
	types := []string{"docker-compose", "docker-swarm"}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			s := &Step{Service: &Service{
				Type:                  typ,
				Name:                  "my-task",
				Config:                "services:\n  web:\n    image: nginx:alpine",
				Timeout:               60,
				ImageRegistry:         "registry.example.com",
				ImageRegistryUsername: "deploy",
				ImageRegistryPassword: "secret",
			}}

			cmd, err := s.buildServiceCommand()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range []string{
				`export PIPELINE_SERVICE_DOCKER_CONFIG=$(mktemp -d)`,
				`export DOCKER_CONFIG="$PIPELINE_SERVICE_DOCKER_CONFIG"`,
				"docker login",
				"--password-stdin",
				`mkdir -p "$PIPELINE_SERVICE_DOCKER_CONFIG/cli-plugins"`,
				`"$_home/.docker/cli-plugins"`,
				`ln -sf "$_p" "$PIPELINE_SERVICE_DOCKER_CONFIG/cli-plugins/$(basename "$_p")"`,
				`[ -n "$PIPELINE_SERVICE_DOCKER_CONFIG" ] && rm -rf "$PIPELINE_SERVICE_DOCKER_CONFIG"`,
			} {
				if !strings.Contains(cmd, want) {
					t.Errorf("command missing %q\ncommand:\n%s", want, cmd)
				}
			}

			if strings.Contains(cmd, "secret") {
				t.Errorf("registry password leaked into command:\n%s", cmd)
			}
		})
	}
}

