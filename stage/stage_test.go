package stage

import (
	"context"
	"testing"

	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/step"
)

func TestStage(t *testing.T) {
	stage := &Stage{
		Name: "test stage",
		Jobs: []*job.Job{
			{
				Name: "test job",
				Steps: []*step.Step{
					{
						Name:    "test step",
						Command: "date",
					},
				},
			},
		},
	}

	if err := stage.Setup("0"); err != nil {
		t.Errorf("Failed to setup stage: %v", err)
	}

	if err := stage.Run(context.Background()); err != nil {
		t.Errorf("Failed to run stage: %v", err)
	}

	if stage.State == nil {
		t.Errorf("Stage state is nil")
	}

	if stage.State.Status != "succeeded" {
		t.Errorf("Stage status is not succeeded")
	}

	if stage.State.StartedAt.IsZero() {
		t.Errorf("Stage started at is zero")
	}

	if stage.State.SucceedAt.IsZero() {
		t.Errorf("Stage succeed at is zero")
	}

	if stage.State.Error != "" {
		t.Errorf("Stage error is not empty")
	}

	if !stage.State.FailedAt.IsZero() {
		t.Errorf("Stage failed at is zero")
	}

	t.Logf("Stage state: %+v", stage.State)
}
