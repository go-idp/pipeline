// Package web embeds the built Pipeline Web frontend (web/ui/dist) into the
// binary, so the `pipeline web` command can serve the management console
// without any external static files.
//
// Build the frontend before building the binary:
//
//	cd web/ui && pnpm install && pnpm build
package web

import "embed"

// Dist is the built frontend.
//
//go:embed ui/dist
var Dist embed.FS
