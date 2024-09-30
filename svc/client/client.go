package client

import (
	"fmt"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/websocket"
)

type Client interface {
	Connect() error
	Close() error
	//
	Run(*pipeline.Pipeline) error
}

type client struct {
	cfg *Config

	core websocket.Client

	done chan error
}

type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("%s(exit code: %d)", e.Message, e.Code)
}

func New(cfg *Config) Client {
	return &client{
		cfg: cfg,
		//
		done: make(chan error),
	}
}
