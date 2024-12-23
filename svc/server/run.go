package server

import (
	"fmt"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/chalk"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/zoox"

	defaults "github.com/go-zoox/zoox/defaults"
)

func (s *server) Run() error {
	if ok := fs.IsExist(s.cfg.Workdir); !ok {
		if err := fs.Mkdirp(s.cfg.Workdir); err != nil {
			return fmt.Errorf("failed to create workdir: %s", err)
		}
	}

	app := defaults.Defaults()

	app.SetBanner(fmt.Sprintf(`
  _____       _______  ___    ___  _          ___         
 / ___/__    /  _/ _ \/ _ \  / _ \(_)__  ___ / (_)__  ___ 
/ (_ / _ \  _/ // // / ___/ / ___/ / _ \/ -_) / / _ \/ -_)
\___/\___/ /___/____/_/    /_/  /_/ .__/\__/_/_/_//_/\__/ 
                                 /_/                      v%s
`, chalk.Green(pipeline.Version)))

	if s.cfg.Username != "" || s.cfg.Password != "" {
		app.Use(func(ctx *zoox.Context) {
			user, pass, ok := ctx.Request.BasicAuth()
			if !ok {
				ctx.Set("WWW-Authenticate", `Basic realm="go-zoox"`)
				ctx.Status(401)
				return
			}

			if !(user == s.cfg.Username && pass == s.cfg.Password) {
				ctx.Status(401)
				return
			}

			ctx.Next()
		})
	}

	err := Mount(app, func(opt *MountConfig) {
		opt.Path = s.cfg.Path
		opt.Workdir = s.cfg.Workdir
		opt.Environment = s.cfg.Environment
	})
	if err != nil {
		return err
	}

	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, map[string]string{
			"version":    pipeline.Version,
			"running_at": app.Runtime().RunningAt().Format("YYYY-MM-DD HH:mm:ss"),
		})
	})

	return app.Run(fmt.Sprintf(":%d", s.cfg.Port))
}
