package commands

import (
	"github.com/go-idp/pipeline/svc/server"
	webui "github.com/go-idp/pipeline/web"
	"github.com/go-zoox/cli"
)

// RegisterWeb registers the `web` command: the full management console.
// It embeds the built frontend (web/ui/dist) and serves it together with the
// server API, i.e. `pipeline web` = frontend (embedded) + backend (server).
//
// The frontend must be built before the binary is built:
//
//	cd web/ui && pnpm install && pnpm build
func RegisterWeb(app *cli.MultipleProgram) {
	app.Register("web", &cli.Command{
		Name:  "web",
		Usage: "the web console of pipeline as a service (embedded frontend + server)",
		Flags: serverFlags(),
		Action: func(ctx *cli.Context) error {
			cfg := serverConfigFromContext(ctx)
			// 嵌入的前端资源（web/ui/dist）
			cfg.WebFS = webui.Dist

			s := server.New(cfg)

			if err := s.Run(); err != nil {
				return err
			}

			return nil
		},
	})
}
