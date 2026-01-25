package server

import (
	"embed"
)

//go:embed console.html
var consoleFS embed.FS

func getConsoleHTML() string {
	data, err := consoleFS.ReadFile("console.html")
	if err != nil {
		return "<html><body><h1>Console not found</h1></body></html>"
	}
	return string(data)
}
