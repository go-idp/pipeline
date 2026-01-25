package stage

import (
	"context"
	"errors"
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
	if s.Timeout > 0 {
		s.logger.Infof("%s[stage(%d/%d): %s] timeout: %d seconds", cfg.Parent, cfg.Current, cfg.Total, s.Name, s.Timeout)
	}
	defer s.logger.Infof("%s[stage(%d/%d): %s] done", cfg.Parent, cfg.Current, cfg.Total, s.Name)

	// Create context with timeout for stage
	var cancel context.CancelFunc
	if s.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s.Timeout)*time.Second)
		defer cancel()
	}

	// job run mode
	//	serial: run jobs in serial => one by one
	//	parallel: run jobs in parallel => all at once
	if s.RunMode == RunModeSerial {
		// serial
		s.logger.Infof("%s[stage(%d/%d): %s] run mode: serial", cfg.Parent, cfg.Current, cfg.Total, s.Name)

		for i, j := range s.Jobs {
			err := j.Run(ctx, func(c *job.RunConfig) {
				c.Total = len(s.Jobs)
				c.Current = i + 1
				c.Parent = fmt.Sprintf("%s[stage(%d/%d): %s]", cfg.Parent, cfg.Current, cfg.Total, s.Name)
			})
			if err != nil {
				s.State.Status = "failed"
				s.State.Error = err.Error()
				s.State.FailedAt = time.Now()
				// Check if error is due to context timeout
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					s.State.Error = fmt.Sprintf("stage timeout after %d seconds: %s", s.Timeout, err.Error())
				}
				return err
			}
		}
	} else {
		// parallel
		s.logger.Infof("%s[stage(%d/%d): %s] run mode: parallel", cfg.Parent, cfg.Current, cfg.Total, s.Name)

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
			// Check if error is due to context timeout
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				s.State.Error = fmt.Sprintf("stage timeout after %d seconds: %s", s.Timeout, err.Error())
			}
			return err
		}
	}

	s.State.Status = "succeeded"
	s.State.SucceedAt = time.Now()

	return nil
}
