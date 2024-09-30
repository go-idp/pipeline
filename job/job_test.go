package job

import (
	"context"
	"testing"

	"github.com/go-idp/pipeline/step"
)

func TestJob(t *testing.T) {
	job := &Job{
		Name: "test job",
		Steps: []*step.Step{
			{
				Name:    "test step 1",
				Command: "date",
			},
			{
				Name:    "test step 2",
				Command: "echo hello",
			},
			{
				Name:    "test step 3",
				Command: "echo world",
			},
		},
	}

	if err := job.Setup("0"); err != nil {
		t.Errorf("Failed to setup job: %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Errorf("Failed to run job: %v", err)
	}

	if job.State == nil {
		t.Errorf("Job state is nil")
	}

	if job.State.Status != "succeeded" {
		t.Errorf("Job status is not succeeded")
	}

	if job.State.StartedAt.IsZero() {
		t.Errorf("Job started at is zero")
	}

	if job.State.SucceedAt.IsZero() {
		t.Errorf("Job succeed at is zero")
	}

	if job.State.Error != "" {
		t.Errorf("Job error is not empty")
	}

	if !job.State.FailedAt.IsZero() {
		t.Errorf("Job failed at is zero")
	}

	t.Logf("Job state: %+v", job.State)
}
