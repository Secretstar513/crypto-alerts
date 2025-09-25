package server

import (
	"html/template"
	"path/filepath"
)

func loadTemplates() *template.Template {
	base := filepath.Join("web", "templates", "base.tmpl.html")
	index := filepath.Join("web", "templates", "index.tmpl.html")
	alerts := filepath.Join("web", "templates", "alerts.tmpl.html")
	channels := filepath.Join("web", "templates", "channels.tmpl.html")
	return template.Must(template.ParseFiles(base, index, alerts, channels))
}
