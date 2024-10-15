package commands

import (
	"fmt"

	"github.com/go-idp/pipeline"
	"github.com/go-idp/pipeline/svc/client"
	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs/type/yaml"
)

func RegisterClient(app *cli.MultipleProgram) {
	app.Register("client", &cli.Command{
		Name:  "client",
		Usage: "the client of pipeline as a service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Specifies the configuration",
				EnvVars:  []string{"CONFIG"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "server",
				Aliases:  []string{"s"},
				Usage:    "Specifies the server",
				EnvVars:  []string{"SERVER"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "username",
				Aliases: []string{"u"},
				Usage:   "Specifies the username",
				EnvVars: []string{"USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "Specifies the password",
				EnvVars: []string{"PASSWORD"},
			},
			&cli.StringFlag{
				Name: "path",
				// Aliases: []string{"p"},
				Usage:   "Specifies the path",
				EnvVars: []string{"SERVER_PATH"},
				Value:   "/",
			},
		},
		Action: func(ctx *cli.Context) error {
			pipeline := pipeline.Pipeline{}
			if configPath := ctx.String("config"); configPath == "" {
				return fmt.Errorf("config is required")
			} else {
				if err := yaml.Read(configPath, &pipeline); err != nil {
					return fmt.Errorf("failed to read config(file: %s): %s", configPath, err)
				}
			}

			cfg := &client.Config{
				Server:   ctx.String("server"),
				Username: ctx.String("username"),
				Password: ctx.String("password"),
				Path:     ctx.String("path"),
			}

			s := client.New(cfg)

			if err := s.Connect(); err != nil {
				return err
			}
			defer s.Close()

			return s.Run(&pipeline)
		},
	})
}
