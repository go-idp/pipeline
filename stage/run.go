package stage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-idp/pipeline/job"
	"golang.org/x/sync/errgroup"
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

// Run runs the stage
func (s *Stage) Run(ctx context.Context, opts ...RunOption) error {
	cfg := &RunConfig{}
	for _, o := range opts {
		o(cfg)
	}

	s.logger.Infof("%s[stage(%d/%d): %s] start", cfg.Parent, cfg.Current, cfg.Total, s.Name)
	defer s.logger.Infof("%s[stage(%d/%d): %s] done", cfg.Parent, cfg.Current, cfg.Total, s.Name)

	g, ctx := errgroup.WithContext(ctx)

	for i, j := range s.Jobs {
		g.Go(func() error {
			return j.Run(ctx, func(c *job.RunConfig) {
				c.Total = len(s.Jobs)
				c.Current = i + 1
				c.Parent = fmt.Sprintf("%s[stage(%d/%d): %s]", cfg.Parent, cfg.Current, cfg.Total, s.Name)
			})
		})
	}

	if err := g.Wait(); err != nil {
		s.State.Status = "failed"
		s.State.Error = err.Error()
		s.State.FailedAt = time.Now()
		return err
	}

	s.State.Status = "succeeded"
	s.State.SucceedAt = time.Now()

	return nil
}
