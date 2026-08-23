package step

import (
	"context"
	"strings"
	"testing"
)

func TestStep(t *testing.T) {
	step := &Step{
		Name:    "test step",
		Command: "date && echo \"CI: $CI\" ",
		Environment: map[string]string{
			"CI": "true",
		},
	}

	if err := step.Setup("0"); err != nil {
		t.Errorf("Failed to setup step: %v", err)
	}

	if err := step.Run(context.Background()); err != nil {
		t.Errorf("Failed to run step: %v", err)
	}

	if step.State == nil {
		t.Errorf("Step state is nil")
	}

	if step.State.Status != "succeeded" {
		t.Errorf("Step status is not succeeded")
	}

	if step.State.StartedAt.IsZero() {
		t.Errorf("Step started at is zero")
	}

	if step.State.SucceedAt.IsZero() {
		t.Errorf("Step succeed at is zero")
	}

	if step.State.Error != "" {
		t.Errorf("Step error is not empty")
	}

	if !step.State.FailedAt.IsZero() {
		t.Errorf("Step failed at is zero")
	}

	// if step.State.ExitCode != 0 {
	// 	t.Errorf("Step exit code is not zero")
	// }

	// if step.State.OOMKilled {
	// 	t.Errorf("Step OOM killed is true")
	// }

	t.Logf("Step state: %+v", step.State)
}

func TestStepSupportServiceTypeDockerCompose(t *testing.T) {
	step := &Step{
		Name: "test step",
		Service: &Service{
			Type: "docker-compose",
			Name: "test_service",
			Config: `
version: '3.7'

services:
  web:
    image: nginx:alpine
    ports:
      - 8080:80
`,
		},
	}

	if err := step.Setup("0"); err != nil {
		t.Errorf("Failed to setup step: %v", err)
	}

	// service steps no longer generate shell commands
	if step.Command != "" {
		t.Errorf("service step should not generate a command, got: %q", step.Command)
	}

	if step.Service.Timeout != 120 {
		t.Errorf("default service timeout mismatch: got %d want 120", step.Service.Timeout)
	}
}

func TestStepSetup_ServiceValidation(t *testing.T) {
	cases := []struct {
		name    string
		service *Service
		wantErr string
	}{
		{
			name:    "missing type",
			service: &Service{Name: "x", Config: "services: {}"},
			wantErr: "service type is required",
		},
		{
			name:    "missing name",
			service: &Service{Type: "docker-compose", Config: "services: {}"},
			wantErr: "service name is required",
		},
		{
			name:    "missing config",
			service: &Service{Type: "docker-compose", Name: "x"},
			wantErr: "service config is required",
		},
		{
			name:    "unsupported type",
			service: &Service{Type: "unknown", Name: "x", Config: "services: {}"},
			wantErr: "unsupported service type",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Step{Name: "s", Service: tc.service}
			err := s.Setup("0")
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}
