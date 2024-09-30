package commands

import (
	"os"
	"strings"

	"github.com/go-idp/pipeline/svc/server"
	"github.com/go-zoox/cli"
)

func RegisterServer(app *cli.MultipleProgram) {
	app.Register("server", &cli.Command{
		Name:  "server",
		Usage: "the server of pipeline as a service",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Specifies the port of server",
				EnvVars: []string{"PORT"},
				Value:   8080,
			},
			&cli.StringFlag{
				Name:    "path",
				Usage:   "Specifies the path of server, web service work path",
				EnvVars: []string{"ENDPOINT", "SERVER_PATH"},
				Value:   "/",
			},
			&cli.StringFlag{
				Name:    "workdir",
				Aliases: []string{"w"},
				Usage:   "Specifies the workdir",
				EnvVars: []string{"WORKDIR"},
				Value:   "/tmp/go-idp/pipeline",
			},
			&cli.StringFlag{
				Name:    "username",
				Aliases: []string{"u"},
				Usage:   "Specifies the username",
				EnvVars: []string{"USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Usage:   "Specifies the password",
				EnvVars: []string{"PASSWORD"},
			},
			&cli.StringSliceFlag{
				Name:    "allow-env",
				Usage:   "Specifies the allowed environment variables",
				EnvVars: []string{"ALLOW_ENV"},
			},
			&cli.BoolFlag{
				Name:    "allow-all-env",
				Usage:   "Specifies the allowed all environment variables",
				EnvVars: []string{"ALLOW_ALL_ENV"},
			},
		},
		Action: func(ctx *cli.Context) error {
			environment := map[string]string{}

			for _, key := range ctx.StringSlice("allow-env") {
				if _, ok := environment[key]; !ok {
					environment[key] = os.Getenv(key)
				}
			}

			if ctx.Bool("allow-all-env") {
				for _, key := range os.Environ() {
					kv := strings.Split(key, "=")
					if len(kv) >= 1 {
						if _, ok := environment[kv[0]]; !ok {
							environment[kv[0]] = kv[1]
						}
					}
				}
			}

			cfg := &server.Config{
				Port: ctx.Int("port"),
				//
				Path: ctx.String("path"),
				//
				Workdir: ctx.String("workdir"),
				//
				Environment: environment,
				//
				Username: ctx.String("username"),
				Password: ctx.String("password"),
			}

			s := server.New(cfg)

			if err := s.Run(); err != nil {
				return err
			}

			return nil
		},
	})
}
