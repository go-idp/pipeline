package step

import (
	"testing"
)

func TestLoadComposeProject_InterpolationForms(t *testing.T) {
	s := &Step{
		Name: "deploy",
		Environment: map[string]string{
			"EUNOMIA_BUILD_ID":      "10000",
			"EUNOMIA_GIT_BRANCH":    "master",
			"EUNOMIA_REGISTRY":      "registry.example.com",
			"UNSET_VAR":             "",
			"EUNOMIA_BUILD_TIMESTAMP": "1727070237",
		},
		Workdir: t.TempDir(),
		Service: &Service{
			Type: "docker-compose",
			Name: "example_task_1234",
			Config: `
version: '3.7'

services:
  web:
    image: ${EUNOMIA_REGISTRY}/nginx:alpine
    environment:
      - BUILD_ID=${EUNOMIA_BUILD_ID}
      - BRANCH=$EUNOMIA_GIT_BRANCH
      - TIMESTAMP=${EUNOMIA_BUILD_TIMESTAMP}
      - LITERAL=$$EUNOMIA_BUILD_ID
      - UNUSED=${UNSET_VAR:-fallback}
    labels:
      - "traefik.port=80"
`,
		},
	}

	project, err := s.loadComposeProject()
	if err != nil {
		t.Fatalf("loadComposeProject() error: %v", err)
	}

	web := project.Services["web"]
	if web.Image != "registry.example.com/nginx:alpine" {
		t.Errorf("image interpolation mismatch: %q", web.Image)
	}

	env := web.Environment
	// ${VAR} form
	if env["BUILD_ID"] == nil || *env["BUILD_ID"] != "10000" {
		t.Errorf("BUILD_ID interpolation mismatch: %v", env["BUILD_ID"])
	}
	// $VAR form
	if env["BRANCH"] == nil || *env["BRANCH"] != "master" {
		t.Errorf("BRANCH interpolation mismatch: %v", env["BRANCH"])
	}
	// $$ escapes to a literal $
	if env["LITERAL"] == nil || *env["LITERAL"] != "$EUNOMIA_BUILD_ID" {
		t.Errorf("$$ escape mismatch: %v", env["LITERAL"])
	}
	// ${VAR:-default} form
	if env["UNUSED"] == nil || *env["UNUSED"] != "fallback" {
		t.Errorf("default-value interpolation mismatch: %v", env["UNUSED"])
	}
}

func TestLoadComposeProject_ProjectNameValidation(t *testing.T) {
	base := &Step{
		Name:    "deploy",
		Workdir: t.TempDir(),
		Service: &Service{
			Type:   "docker-compose",
			Config: "version: '3.7'\nservices:\n  web:\n    image: nginx:alpine",
		},
	}

	// valid lowercase name with dashes/underscores
	ok := *base
	ok.Service = &Service{Type: "docker-compose", Name: "eunomia-task-10086", Config: base.Service.Config}
	if _, err := ok.loadComposeProject(); err != nil {
		t.Errorf("valid project name should load, got: %v", err)
	}

	// invalid (uppercase) name must fail like `docker compose`
	bad := *base
	bad.Service = &Service{Type: "docker-compose", Name: "MyProject", Config: base.Service.Config}
	if _, err := bad.loadComposeProject(); err == nil {
		t.Errorf("uppercase project name should be rejected")
	}
}

func TestLoadComposeProject_DependsOnAndPorts(t *testing.T) {
	s := &Step{
		Name:    "deploy",
		Workdir: t.TempDir(),
		Service: &Service{
			Type: "docker-compose",
			Name: "example",
			Config: `
version: '3.7'

services:
  db:
    image: postgres:13
    environment:
      POSTGRES_PASSWORD: secret
    volumes:
      - db-data:/var/lib/postgresql/data
  web:
    image: nginx:alpine
    depends_on:
      - db
    ports:
      - "8080:80"

volumes:
  db-data:
`,
		},
	}

	project, err := s.loadComposeProject()
	if err != nil {
		t.Fatalf("loadComposeProject() error: %v", err)
	}

	if len(project.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(project.Services))
	}

	web := project.Services["web"]
	if len(web.DependsOn) != 1 {
		t.Errorf("depends_on not parsed: %v", web.DependsOn)
	}
	if _, ok := web.DependsOn["db"]; !ok {
		t.Errorf("depends_on should reference db: %v", web.DependsOn)
	}

	db := project.Services["db"]
	if db.Environment["POSTGRES_PASSWORD"] == nil || *db.Environment["POSTGRES_PASSWORD"] != "secret" {
		t.Errorf("db env mismatch: %v", db.Environment["POSTGRES_PASSWORD"])
	}

	if _, ok := project.Volumes["db-data"]; !ok {
		t.Errorf("named volume db-data not parsed")
	}
}

func TestLoadComposeProject_EmptyConfig(t *testing.T) {
	s := &Step{
		Name:    "deploy",
		Workdir: t.TempDir(),
		Service: &Service{Type: "docker-compose", Name: "example", Config: ""},
	}

	if _, err := s.loadComposeProject(); err == nil {
		t.Fatalf("expected error for empty config")
	}
}

func TestLoadComposeProject_InvalidYAML(t *testing.T) {
	s := &Step{
		Name:    "deploy",
		Workdir: t.TempDir(),
		Service: &Service{Type: "docker-compose", Name: "example", Config: "services: [unclosed"},
	}

	if _, err := s.loadComposeProject(); err == nil {
		t.Fatalf("expected error for invalid YAML")
	}
}

func TestLoadComposeProject_UnsupportedTopLevelKey(t *testing.T) {
	s := &Step{
		Name:    "deploy",
		Workdir: t.TempDir(),
		Service: &Service{
			Type: "docker-compose",
			Name: "example",
			Config: `
version: '3.7'
totally_unknown_key: true
services:
  web:
    image: nginx:alpine
`,
		},
	}

	if _, err := s.loadComposeProject(); err == nil {
		t.Fatalf("expected error for unsupported top-level key")
	}
}
