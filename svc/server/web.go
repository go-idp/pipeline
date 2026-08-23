package server

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/go-zoox/headers"
	"github.com/go-zoox/zoox"
)

// MountWeb serves the embedded SPA (web/ui/dist) on the application root.
//
// It uses a middleware so that:
//   - /api/* and the WebSocket path pass through to their own handlers;
//   - asset files (with an extension) are served from the embedded fs;
//   - any other path falls back to index.html, enabling client-side routing.
//
// It is used by the `pipeline web` command only; the lightweight `pipeline
// server` command keeps the JSON version response on "/".
func MountWeb(app *zoox.Application, fsys fs.FS, wsPath string) error {
	sub, err := fs.Sub(fsys, "ui/dist")
	if err != nil {
		return err
	}

	app.Use(func(ctx *zoox.Context) {
		if ctx.Method != http.MethodGet && ctx.Method != http.MethodHead {
			ctx.Next()
			return
		}

		p := ctx.Path
		if strings.HasPrefix(p, "/api/") || p == wsPath {
			ctx.Next()
			return
		}

		name := strings.TrimPrefix(p, "/")
		if name == "" || name == "." {
			name = "index.html"
		}

		serveFile(ctx, sub, name)
	})

	return nil
}

// serveFile writes a file from the embedded fs to the response, falling back
// to index.html for SPA routes (paths without a file extension).
func serveFile(ctx *zoox.Context, sub fs.FS, name string) {
	if path.Ext(name) == "" {
		// client-side route: always serve index.html
		writeFile(ctx, sub, "index.html", "text/html; charset=utf-8")
		return
	}

	contentType := mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	writeFile(ctx, sub, name, contentType)
}

func writeFile(ctx *zoox.Context, sub fs.FS, name string, contentType string) {
	data, err := fs.ReadFile(sub, name)
	if err != nil {
		// 文件不存在：非 SPA 路由（带扩展名）直接 404
		ctx.Status(404)
		return
	}

	ctx.SetHeader(headers.ContentType, contentType)
	ctx.Write(data)
}
