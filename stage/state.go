package stage

import "time"

type State struct {
	ID     string `json:"id" yaml:"id"`
	Status string `json:"status" yaml:"status"` // pending | running | succeeded | failed
	//
	StartedAt time.Time `json:"started_at" yaml:"started_at"`
	SucceedAt time.Time `json:"succeed_at" yaml:"succeed_at"`
	FailedAt  time.Time `json:"failed_at" yaml:"failed_at"`
	//
	Error string `json:"error" yaml:"error"`
}
