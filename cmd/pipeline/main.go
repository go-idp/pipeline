package main

import (
	"github.com/go-idp/pipeline"
	"github.com/go-idp/pipeline/cmd/pipeline/commands"
	"github.com/go-zoox/cli"
)

func main() {
	app := cli.NewMultipleProgram(&cli.MultipleProgramConfig{
		Name:    "pipeline",
		Usage:   "single is a program that has a single command.",
		Version: pipeline.Version,
	})

	commands.RegisterRun(app)

	commands.RegisterServer(app)
	commands.RegisterClient(app)

	app.Run()
}
