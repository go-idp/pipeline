package client

import (
	"fmt"
	"io"
	"os"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/websocket"
)

type Client interface {
	Connect() error
	Close() error
	//
	Run(*pipeline.Pipeline) error
	//
	SetStdout(stdout io.Writer)
	SetStderr(stderr io.Writer)
}

type client struct {
	cfg *Config

	core websocket.Client

	done chan error

	stdout io.Writer
	stderr io.Writer
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
		stdout: os.Stdout,
		stderr: os.Stderr,
		//
		done: make(chan error),
	}
}

func (c *client) SetStdout(stdout io.Writer) {
	c.stdout = stdout

	if c.stderr == nil {
		c.stderr = stdout
	}
}

func (c *client) SetStderr(stderr io.Writer) {
	c.stderr = stderr
}
