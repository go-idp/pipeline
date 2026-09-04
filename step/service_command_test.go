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
// running with HOME=/root), Docker tried to write ~/.docker/config.json and
// failed with "error saving credentials: mkdir /root: read-only file system",
// aborting the whole deploy. The command must point DOCKER_CONFIG at a writable
// temp dir before logging in, and the same DOCKER_CONFIG stays exported for the
// following deploy (--with-registry-auth / compose up) so auth is still sent.
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

