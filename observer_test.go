package pipeline

import (
	"context"
	"sync"
	"testing"

	"github.com/go-idp/pipeline/event"
	"github.com/go-idp/pipeline/job"
	"github.com/go-idp/pipeline/stage"
	"github.com/go-idp/pipeline/step"
)

// TestObserver verifies that the run state observer receives events from all
// four levels in a valid order.
func TestObserver(t *testing.T) {
	pl := &Pipeline{
		Name: "observer-test",
		Stages: []*stage.Stage{
			{
				Name: "build",
				Jobs: []*job.Job{
					{
						Name: "job1",
						Steps: []*step.Step{
							{Name: "step1", Command: "echo hello"},
							{Name: "step2", Command: "echo world"},
						},
					},
				},
			},
		},
	}

	var mu sync.Mutex
	events := make([]event.Event, 0)
	pl.SetObserve(func(ev event.Event) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
	})

	if err := pl.Run(context.Background()); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(events) == 0 {
		t.Fatalf("expected events, got none")
	}

	countByLevel := map[event.Level]int{}
	for _, ev := range events {
		countByLevel[ev.Level]++
		if ev.Status != "running" && ev.Status != "succeeded" {
			t.Errorf("unexpected status %q for level %s", ev.Status, ev.Level)
		}
		if ev.ID == "" {
			t.Errorf("event id is empty for level %s", ev.Level)
		}
		if ev.Name == "" {
			t.Errorf("event name is empty for level %s", ev.Level)
		}
	}

	if countByLevel[event.LevelPipeline] != 2 {
		t.Errorf("pipeline events: got %d, want 2 (running + succeeded)", countByLevel[event.LevelPipeline])
	}
	if countByLevel[event.LevelStage] != 2 {
		t.Errorf("stage events: got %d, want 2 (running + succeeded)", countByLevel[event.LevelStage])
	}
	if countByLevel[event.LevelJob] != 2 {
		t.Errorf("job events: got %d, want 2 (running + succeeded)", countByLevel[event.LevelJob])
	}
	if countByLevel[event.LevelStep] != 4 {
		t.Errorf("step events: got %d, want 4 (2 steps x running + succeeded)", countByLevel[event.LevelStep])
	}

	// the last event must be the pipeline succeeded event
	last := events[len(events)-1]
	if last.Level != event.LevelPipeline || last.Status != "succeeded" {
		t.Errorf("last event should be pipeline/succeeded, got %s/%s", last.Level, last.Status)
	}
}
