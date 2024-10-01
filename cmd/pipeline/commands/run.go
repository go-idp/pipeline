package commands

import (
	"context"
	"os"
	"strings"

	"github.com/go-zoox/core-utils/fmt"
	"github.com/go-zoox/core-utils/regexp"
	"github.com/go-zoox/debug"
	"github.com/go-zoox/fetch"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs/type/yaml"
)

func RegisterRun(app *cli.MultipleProgram) {
	app.Register("run", &cli.Command{
		Name:  "run",
		Usage: "run a pipeline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Specifies the configuration",
				EnvVars: []string{"PIPELINE_CONFIG"},
				// Required: true,
				Value: findConfig(),
			},
			&cli.StringFlag{
				Name:    "workdir",
				Aliases: []string{"w"},
				Usage:   "Specifies the workdir",
				EnvVars: []string{"PIPELINE_WORKDIR"},
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
			fmt.Fprintf(os.Stdout, `
  _____       _______  ___    ___  _          ___         
 / ___/__    /  _/ _ \/ _ \  / _ \(_)__  ___ / (_)__  ___ 
/ (_ / _ \  _/ // // / ___/ / ___/ / _ \/ -_) / / _ \/ -_)
\___/\___/ /___/____/_/    /_/  /_/ .__/\__/_/_/_//_/\__/ 
                                 /_/                      v%s
`+"\n\n", pipeline.Version)
			config := ctx.String("config")
			if config == "" {
				return fmt.Errorf("config is required")
			}

			// support for remote config
			if ok := regexp.Match(`^https?://`, config); ok {
				url := config
				config = fs.TmpFilePath() + ".yaml"
				response, err := fetch.Get(url)
				if err != nil {
					return fmt.Errorf("failed to fetch config(url: %s): %s", url, err)
				}

				if err := fs.WriteFile(config, []byte(response.String())); err != nil {
					return fmt.Errorf("failed to write config(file: %s): %s", config, err)
				}

				if !debug.IsDebugMode() {
					defer fs.RemoveFile(config)
				} else {
					logger.Infof("load config from %s to %s", url, config)
				}
			}

			p := &pipeline.Pipeline{}
			if err := yaml.Read(config, p); err != nil {
				return fmt.Errorf("failed to read config(file: %s): %s", config, err)
			}

			if workdir := ctx.String("workdir"); workdir != "" {
				// pl.Workdir = workdir
				p.SetWorkdir(workdir)
			}

			if len(ctx.StringSlice("allow-env")) != 0 {
				environment := map[string]string{}
				for _, key := range ctx.StringSlice("allow-env") {
					environment[key] = os.Getenv(key)
				}
				p.SetEnvironment(environment)
			}

			if ctx.Bool("allow-all-env") {
				for _, key := range os.Environ() {
					kv := strings.Split(key, "=")
					if len(kv) >= 1 {
						if _, ok := p.Environment[kv[0]]; !ok {
							p.Environment[kv[0]] = kv[1]
						}
					}
				}
			}

			if debug.IsDebugMode() {
				fmt.PrintJSON(p)
			}

			return p.Run(context.Background())
		},
	})
}

func findConfig() string {
	// @1 .pipeline.yaml
	if ok := fs.IsExist(".pipeline.yaml"); ok {
		return ".pipeline.yaml"
	}

	// @2 .go-idp/pipeline.yaml
	if ok := fs.IsExist(".go-idp/pipeline.yaml"); ok {
		return ".go-idp/pipeline.yaml"
	}

	return ""
}
