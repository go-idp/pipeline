package client

import (
	"fmt"

	"github.com/go-idp/pipeline"
	"github.com/go-idp/pipeline/svc/action"
)

func (c *client) Run(p *pipeline.Pipeline) error {
	if c.core == nil {
		return fmt.Errorf("client is not connected")
	}

	msg, err := action.Run.Encode(p)
	if err != nil {
		return fmt.Errorf("failed to encode run action: %s", err)
	}

	if err := c.core.SendTextMessage(msg); err != nil {
		return fmt.Errorf("failed to send run action: %s", err)
	}

	return <-c.done
}
