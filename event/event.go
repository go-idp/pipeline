package event

import "time"

// Level is the execution level of a pipeline run event.
type Level string

const (
	// LevelPipeline is the pipeline level.
	LevelPipeline Level = "pipeline"
	// LevelStage is the stage level.
	LevelStage Level = "stage"
	// LevelJob is the job level.
	LevelJob Level = "job"
	// LevelStep is the step level.
	LevelStep Level = "step"
)

// Event is a run state change event emitted by the pipeline execution
// engine (pipeline / stage / job / step).
type Event struct {
	// Level is the execution level.
	Level Level `json:"level"`
	// ID is the hierarchical state id, e.g. "<run-id>.0.1.2".
	ID string `json:"id"`
	// Name is the name of the node.
	Name string `json:"name"`
	// Status is pending | running | succeeded | failed | cancelled.
	Status string `json:"status"`
	// Error is the error message when the node failed.
	Error string `json:"error,omitempty"`
	// StartedAt is when the node was set up.
	StartedAt time.Time `json:"started_at"`
	// EndedAt is when the node reached a terminal state.
	EndedAt *time.Time `json:"ended_at,omitempty"`
}

// Observer receives run state change events. It is optional and safe to
// leave nil; the execution engine never blocks on it.
type Observer func(Event)
