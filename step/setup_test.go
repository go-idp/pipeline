package step

import (
	"encoding/base64"
	"testing"
)

func TestStepSetup_Defaults(t *testing.T) {
	s := &Step{
		Name:    "s",
		Command: "echo hi",
	}

	if err := s.Setup("sid"); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	if s.State == nil {
		t.Fatalf("state is nil")
	}
	if s.State.ID != "sid" {
		t.Fatalf("state id mismatch: got %q", s.State.ID)
	}
	if s.State.Status != "running" {
		t.Fatalf("state status mismatch: got %q", s.State.Status)
	}
	if s.State.StartedAt.IsZero() {
		t.Fatalf("started_at is zero")
	}

	if s.Environment == nil {
		t.Fatalf("environment should be initialized")
	}
	if s.Timeout != 86400 {
		t.Fatalf("default timeout mismatch: got %d want %d", s.Timeout, 86400)
	}
}

func TestStepSetup_LanguageAndPluginConflict(t *testing.T) {
	s := &Step{
		Name:    "s",
		Command: "echo hi",
		Language: &Language{
			Name:    "node",
			Version: "20",
		},
		Plugin: &Plugin{
			Image: "alpine:3",
		},
	}

	if err := s.Setup("sid"); err == nil {
		t.Fatalf("expected error when language and plugin are both set")
	}
}

func TestStepSetup_PluginSettingsEnvInjection(t *testing.T) {
	originCommand := "echo hello"
	s := &Step{
		Name:    "s",
		Command: originCommand,
		Environment: map[string]string{
			"ENV": "x",
		},
		Plugin: &Plugin{
			Image:       "plugin:image",
			Settings:    map[string]string{"key": "value", "from_env": "${ENV}"},
			inheritEnv:  true,
			Entrypoint:  "",
			ImageRegistry:         "reg",
			ImageRegistryUsername: "u",
			ImageRegistryPassword: "p",
		},
	}

	if err := s.Setup("sid"); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	if s.Image != "plugin:image" {
		t.Fatalf("image should be plugin image: got %q", s.Image)
	}
	if s.Plugin.Entrypoint != "/pipeline/plugin/run" {
		t.Fatalf("plugin entrypoint default mismatch: got %q", s.Plugin.Entrypoint)
	}
	if s.Environment == nil {
		t.Fatalf("environment is nil")
	}
	if got := s.Environment["ENV"]; got != "x" {
		t.Fatalf("expected inherited ENV=x, got %q", got)
	}
	if got := s.Environment["PIPELINE_PLUGIN_COMMAND"]; got != base64.StdEncoding.EncodeToString([]byte(originCommand)) {
		t.Fatalf("PIPELINE_PLUGIN_COMMAND mismatch: got %q", got)
	}
	if got := s.Environment["PIPELINE_PLUGIN_SETTINGS_KEY"]; got != "value" {
		t.Fatalf("PIPELINE_PLUGIN_SETTINGS_KEY mismatch: got %q", got)
	}
	if got := s.Environment["PIPELINE_PLUGIN_SETTINGS_FROM_ENV"]; got != "x" {
		t.Fatalf("PIPELINE_PLUGIN_SETTINGS_FROM_ENV mismatch: got %q", got)
	}
}

