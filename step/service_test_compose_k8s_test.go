package step

import (
	"strings"
	"testing"
)

func TestLoadComposeProject(t *testing.T) {
	s := &Step{
		Name: "deploy",
		Environment: map[string]string{
			"EUNOMIA_BUILD_TIMESTAMP": "1727070237",
			"CI":                      "true",
		},
		Workdir: t.TempDir(),
		Service: &Service{
			Type: "docker-compose",
			Name: "example_task_1234",
			Config: `
version: '3.7'

services:
  web:
    image: nginx:alpine
    environment:
      BUILD_TIMESTAMP: $EUNOMIA_BUILD_TIMESTAMP
    ports:
      - "8080:80"
`,
		},
	}

	project, err := s.loadComposeProject()
	if err != nil {
		t.Fatalf("loadComposeProject() error: %v", err)
	}

	if project.Name != "example_task_1234" {
		t.Errorf("project name mismatch: got %q want %q", project.Name, "example_task_1234")
	}

	if len(project.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(project.Services))
	}

	web, ok := project.Services["web"]
	if !ok {
		t.Fatalf("service web not found")
	}

	// env interpolation: $EUNOMIA_BUILD_TIMESTAMP resolved from step environment
	if web.Environment["BUILD_TIMESTAMP"] == nil || *web.Environment["BUILD_TIMESTAMP"] != "1727070237" {
		t.Errorf("env interpolation mismatch: got %v want %q", web.Environment["BUILD_TIMESTAMP"], "1727070237")
	}

	if web.Image != "nginx:alpine" {
		t.Errorf("image mismatch: got %q", web.Image)
	}
}

func TestLoadComposeProject_InvalidConfig(t *testing.T) {
	s := &Step{
		Name:  "deploy",
		Workdir: t.TempDir(),
		Service: &Service{
			Type: "docker-compose",
			Name: "example",
			Config: `
services:
  web:
    ports:
      - "not-a-valid-port"
`,
		},
	}

	if _, err := s.loadComposeProject(); err == nil {
		t.Fatalf("expected error for invalid compose config")
	}
}

func TestDecodeK8sManifests(t *testing.T) {
	manifest := `
apiVersion: v1
kind: Namespace
metadata:
  name: demo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
        - name: web
          image: nginx:alpine
---
# empty document should be skipped
---
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: demo
`

	objects, err := decodeK8sManifests(manifest)
	if err != nil {
		t.Fatalf("decodeK8sManifests() error: %v", err)
	}

	if len(objects) != 3 {
		t.Fatalf("expected 3 objects, got %d", len(objects))
	}

	kinds := []string{}
	for _, o := range objects {
		kinds = append(kinds, o.GetKind())
	}

	if kinds[0] != "Namespace" || kinds[1] != "Deployment" || kinds[2] != "Service" {
		t.Errorf("kinds mismatch: %v", kinds)
	}
}

func TestDecodeK8sManifests_Invalid(t *testing.T) {
	if _, err := decodeK8sManifests("this is not valid yaml: [::"); err == nil {
		t.Fatalf("expected error for invalid manifest")
	}

	if _, err := decodeK8sManifests(""); err == nil || !strings.Contains(err.Error(), "no k8s objects") {
		t.Fatalf("expected 'no k8s objects' error, got %v", err)
	}
}
