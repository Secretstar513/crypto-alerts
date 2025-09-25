package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.Index)
	r.Post("/alerts", h.CreateAlert)
	r.Post("/alerts/{id}/toggle", h.ToggleAlert)
	r.Post("/alerts/{id}/delete", h.DeleteAlert)

	r.Get("/channels", h.ChannelsPage)
	r.Post("/channels/email", h.UpsertEmail)
	r.Post("/channels/telegram", h.UpsertTelegram)

	fs := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	return r
}
