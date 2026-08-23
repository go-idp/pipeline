package commands

import (
	"os"
	"strings"

	"github.com/go-idp/pipeline/svc/server"
	"github.com/go-zoox/cli"
)

// serverFlags 是 server / web 两个命令共享的命令行参数。
func serverFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "Specifies the port of server",
			EnvVars: []string{"PORT"},
			Value:   8080,
		},
		&cli.StringFlag{
			Name:    "path",
			Usage:   "Specifies the websocket path of server",
			EnvVars: []string{"ENDPOINT", "SERVER_PATH"},
			Value:   "/ws",
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
		&cli.IntFlag{
			Name:    "max-concurrent",
			Usage:   "Specifies the maximum concurrent pipeline executions",
			EnvVars: []string{"MAX_CONCURRENT"},
			Value:   2,
		},
		&cli.IntFlag{
			Name:    "task-timeout",
			Usage:   "Specifies the default timeout (seconds) for pipelines without an explicit timeout, 0 disables it",
			EnvVars: []string{"TASK_TIMEOUT"},
			Value:   3600,
		},
		&cli.StringFlag{
			Name:    "task-executor",
			Usage:   "Specifies how tasks are executed: in-process (default) or subprocess (isolated pipeline run process)",
			EnvVars: []string{"TASK_EXECUTOR"},
			Value:   "in-process",
		},
	}
}

// serverConfigFromContext 从命令行上下文构建 server 配置（含环境变量白名单）。
func serverConfigFromContext(ctx *cli.Context) *server.Config {
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

	return &server.Config{
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
		//
		MaxConcurrent: ctx.Int("max-concurrent"),
		//
		TaskTimeout: ctx.Int64("task-timeout"),
		//
		TaskExecutor: ctx.String("task-executor"),
	}
}

// RegisterServer registers the `server` command: a lightweight API-only
// service (REST + WebSocket) without the embedded web console.
func RegisterServer(app *cli.MultipleProgram) {
	app.Register("server", &cli.Command{
		Name:  "server",
		Usage: "the server of pipeline as a service (api only)",
		Flags: serverFlags(),
		Action: func(ctx *cli.Context) error {
			cfg := serverConfigFromContext(ctx)

			s := server.New(cfg)

			if err := s.Run(); err != nil {
				return err
			}

			return nil
		},
	})
}
