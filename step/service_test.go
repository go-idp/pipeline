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
