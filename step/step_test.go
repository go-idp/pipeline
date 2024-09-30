package step

import (
	"context"
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
