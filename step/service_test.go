package step

import (
	"testing"
)

func TestNormalizeRegistryHost(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"registry.example.com", "registry.example.com"},
		{"https://registry.example.com", "registry.example.com"},
		{"http://registry.example.com:5000/", "registry.example.com:5000"},
		{"  https://registry.example.com/  ", "registry.example.com"},
	}

	for _, tc := range cases {
		if got := normalizeRegistryHost(tc.in); got != tc.want {
			t.Errorf("normalizeRegistryHost(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRegistryAuthKeys(t *testing.T) {
	keys := registryAuthKeys("https://registry.example.com:5000")
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %v", keys)
	}
	if keys[0] != "https://registry.example.com:5000" {
		t.Errorf("raw key mismatch: %q", keys[0])
	}
	if keys[1] != "registry.example.com:5000" {
		t.Errorf("host key mismatch: %q", keys[1])
	}

	// docker hub gets the index key too
	keys = registryAuthKeys("docker.io")
	found := false
	for _, k := range keys {
		if k == "https://index.docker.io/v1/" {
			found = true
		}
	}
	if !found {
		t.Errorf("docker hub keys should include https://index.docker.io/v1/, got %v", keys)
	}
}

func TestStepSetup_ServiceValidTypes(t *testing.T) {
	for _, serviceType := range []string{"docker-compose", "docker-swarm", "kubernetes", "k8s"} {
		s := &Step{
			Name: "s",
			Service: &Service{
				Type:   serviceType,
				Name:   "stack-x",
				Config: "services: {}",
			},
		}

		if err := s.Setup("0"); err != nil {
			t.Errorf("Setup() for type %q error: %v", serviceType, err)
		}
	}
}

func TestStepSetup_ServiceEnvInterpolation(t *testing.T) {
	s := &Step{
		Name: "s",
		Environment: map[string]string{
			"EUNOMIA_TASK_ID":   "10086",
			"REGISTRY_USERNAME": "deploy-user",
			"REGISTRY_PASSWORD": "s3cret",
			"K8S_NAMESPACE":     "demo",
		},
		Service: &Service{
			Type:                  "docker-compose",
			Name:                  "task-${EUNOMIA_TASK_ID}",
			Config:                "services: {}",
			ImageRegistry:         "https://registry.example.com/",
			ImageRegistryUsername: "${REGISTRY_USERNAME}",
			ImageRegistryPassword: "${REGISTRY_PASSWORD}",
		},
	}

	if err := s.Setup("0"); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	if s.Service.Name != "task-10086" {
		t.Errorf("name interpolation mismatch: %q", s.Service.Name)
	}
	if s.Service.ImageRegistryUsername != "deploy-user" {
		t.Errorf("registry username interpolation mismatch: %q", s.Service.ImageRegistryUsername)
	}
	if s.Service.ImageRegistryPassword != "s3cret" {
		t.Errorf("registry password interpolation mismatch: %q", s.Service.ImageRegistryPassword)
	}
	// non-${VAR} values are untouched
	if s.Service.ImageRegistry != "https://registry.example.com/" {
		t.Errorf("registry url should not be modified: %q", s.Service.ImageRegistry)
	}
}

func TestStepSetup_ServiceEnvInterpolationMissingVar(t *testing.T) {
	s := &Step{
		Name: "s",
		Service: &Service{
			Type:   "docker-compose",
			Name:   "task-${UNSET_VAR}",
			Config: "services: {}",
		},
	}

	if err := s.Setup("0"); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	// missing vars are left as-is
	if s.Service.Name != "task-${UNSET_VAR}" {
		t.Errorf("missing var should be left as-is, got %q", s.Service.Name)
	}
}
