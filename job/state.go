package job

import "time"

type State struct {
	ID     string `yaml:"id"`
	Status string `yaml:"status"` // pending | running | succeeded | failed
	//
	StartedAt time.Time `yaml:"started_at"`
	SucceedAt time.Time `yaml:"succeed_at"`
	FailedAt  time.Time `yaml:"failed_at"`
	//
	Error string `yaml:"error"`
}
