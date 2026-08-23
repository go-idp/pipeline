package examples_test

import (
	"os"
	"testing"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/encoding/yaml"
)

// TestExampleConfigsParse ensures every example YAML file still parses and
// decodes into the pipeline model, and that service examples use the native
// Service configuration.
func TestExampleConfigsParse(t *testing.T) {
	files := []string{
		"basic.yml",
		"checkout-build-deploy.yml",
		"docker.yaml",
		"github.yaml",
		"language.yml",
		"plugin.yml",
		"run.yaml",
		"step-engine.yaml",
		"step-engine-ssh.yaml",
		"step-service-docker-compose.yaml",
		"step-service-docker-swarm.yaml",
		"step-service-kubernetes.yaml",
		"service-deploy.yml",
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}

		var p pipeline.Pipeline
		if err := yaml.Decode(data, &p); err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}

		t.Logf("%s: ok (%d stages)", f, len(p.Stages))
	}
}

// TestExampleServiceSteps ensures service examples decode into the native
// Service struct with type and config populated.
func TestExampleServiceSteps(t *testing.T) {
	files := []string{
		"step-service-docker-compose.yaml",
		"step-service-docker-swarm.yaml",
		"step-service-kubernetes.yaml",
		"service-deploy.yml",
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}

		var p pipeline.Pipeline
		if err := yaml.Decode(data, &p); err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}

		found := false
		for _, stage := range p.Stages {
			for _, job := range stage.Jobs {
				for _, step := range job.Steps {
					if step.Service != nil {
						found = true
						if step.Service.Type == "" || step.Service.Config == "" {
							t.Fatalf("%s: service step missing type/config", f)
						}
					}
				}
			}
		}
		if !found {
			t.Fatalf("%s: no service step found", f)
		}
	}
}
