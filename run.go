package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-idp/pipeline/stage"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/uuid"
)

// Run runs the pipeline (workflow runner)
func (p *Pipeline) Run(ctx context.Context, opts ...RunOption) error {
	cfg := &RunConfig{
		ID: uuid.V4(),
	}
	for _, o := range opts {
		o(cfg)
	}

	//
	logger.Infof("[workflow] start to run (name: %s)", p.Name)
	defer func() {
		logger.Infof("[workflow] done to run (name: %s, workdir: %s)", p.Name, p.Workdir)
	}()

	if err := p.prepare(cfg.ID); err != nil {
		return err
	}
	defer p.clean()

	plog := p.getLogger()
	plog.Infof("[workflow] start")
	plog.Infof("[workflow] version: %s", Version)
	plog.Infof("[workflow] name: %s", p.Name)
	plog.Infof("[workflow] workdir: %s", p.Workdir)
	plog.Infof("[workflow] timeout: %d seconds", p.Timeout)
	defer plog.Infof("[workflow] done")

	// Create context with timeout for pipeline
	var cancel context.CancelFunc
	if p.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(p.Timeout)*time.Second)
		defer cancel()
	}

	for i, s := range p.Stages {
		err := s.Run(ctx, func(cfg *stage.RunConfig) {
			cfg.Total = len(p.Stages)
			cfg.Current = i + 1
		})

		if err != nil {
			p.State.Status = "failed"
			p.State.Error = err.Error()
			p.State.FailedAt = time.Now()
			// Check if error is due to context timeout
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				p.State.Error = fmt.Sprintf("pipeline timeout after %d seconds: %s", p.Timeout, err.Error())
			}
			return err
		}
	}

	p.State.Status = "succeeded"
	p.State.SucceedAt = time.Now()

	return nil
}
