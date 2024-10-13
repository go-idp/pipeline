package job

import (
	"context"
	"fmt"
	"time"

	"github.com/go-idp/pipeline/step"
)

// RunConfig is the config for run
type RunConfig struct {
	// Total is the total count of the parent steps
	Total int
	// Current is the current index of the parent steps
	Current int
	// Parent is the parent name
	Parent string
}

// RunOption is the option for run
type RunOption func(cfg *RunConfig)

// Run runs the job
func (j *Job) Run(ctx context.Context, opts ...RunOption) error {
	cfg := &RunConfig{}
	for _, o := range opts {
		o(cfg)
	}

	j.logger.Infof("%s[job(%d/%d): %s] start", cfg.Parent, cfg.Current, cfg.Total, j.Name)
	defer j.logger.Infof("%s[job(%d/%d): %s] done", cfg.Parent, cfg.Current, cfg.Total, j.Name)

	for i, s := range j.Steps {
		err := s.Run(ctx, func(c *step.RunConfig) {
			c.Total = len(j.Steps)
			c.Current = i + 1
			c.Parent = fmt.Sprintf("%s[job(%d/%d): %s]", cfg.Parent, cfg.Current, cfg.Total, j.Name)
		})

		if err != nil {
			j.State.Status = "failed"
			j.State.Error = err.Error()
			j.State.FailedAt = time.Now()
			return err
		}
	}

	j.State.Status = "succeeded"
	j.State.SucceedAt = time.Now()

	return nil
}
