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
		"docker version --format '{{.Server.Version}}'",
		"17.05",
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
		{"27.3.1", 17, 5, true},
		{"24.0.7", 17, 5, true},
		{"18.09.7", 17, 5, true},
		{"17.05.0-ce", 17, 5, true},
		{"17.06.2", 17, 5, true},
		{"17.04.0", 17, 5, false},
		{"16.04.0", 17, 5, false},
		{"1.13.1", 17, 5, false},
		{"", 17, 5, false},
		{"abc", 17, 5, false},
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
