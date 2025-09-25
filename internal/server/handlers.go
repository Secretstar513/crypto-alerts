package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/Secretstar513/crypto-alerts/internal/app"
	"github.com/Secretstar513/crypto-alerts/internal/domain"
	"github.com/Secretstar513/crypto-alerts/internal/notif"
)

type Handlers struct {
	App  *app.App
	tpl  *template.Template
}

func NewHandlers(a *app.App) *Handlers {
	return &Handlers{App: a, tpl: loadTemplates()}
}

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	list, _ := h.App.ListAlerts()
	data := map[string]any{"Alerts": list, "Page": "alerts"}
	_ = h.tpl.ExecuteTemplate(w, "index", data)
}

func (h *Handlers) CreateAlert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400); return
	}
	symbol := r.FormValue("symbol")
	dir := r.FormValue("direction")
	thr, _ := strconv.ParseFloat(r.FormValue("threshold"), 64)
	al, err := h.App.CreateAlert(symbol, thr, domain.Direction(dir))
	if err != nil {
		http.Error(w, err.Error(), 400); return
	}
	log.Info().Str("id", al.ID).Msg("alert created")

	// return partial alerts list for htmx replace
	list, _ := h.App.ListAlerts()
	w.Header().Set("HX-Trigger", "alert-changed")
	_ = h.tpl.ExecuteTemplate(w, "alerts", map[string]any{"Alerts": list})
}

func (h *Handlers) ToggleAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	enable := r.FormValue("enable") == "true"
	if err := h.App.ToggleAlert(id, enable); err != nil {
		http.Error(w, err.Error(), 400); return
	}
	list, _ := h.App.ListAlerts()
	_ = h.tpl.ExecuteTemplate(w, "alerts", map[string]any{"Alerts": list})
}

func (h *Handlers) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.App.DeleteAlert(id); err != nil {
		http.Error(w, err.Error(), 400); return
	}
	list, _ := h.App.ListAlerts()
	_ = h.tpl.ExecuteTemplate(w, "alerts", map[string]any{"Alerts": list})
}

func (h *Handlers) ChannelsPage(w http.ResponseWriter, r *http.Request) {
	ch, _ := h.App.ListChannels()
	data := map[string]any{"Channels": ch, "Page": "channels"}
	_ = h.tpl.ExecuteTemplate(w, "channels", data)
}

func (h *Handlers) UpsertEmail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cfg := notif.EmailConfig{
		Host: r.FormValue("host"),
		Port: r.FormValue("port"),
		User: r.FormValue("user"),
		Pass: r.FormValue("pass"),
		From: r.FormValue("from"),
		To:   r.FormValue("to"),
	}
	enabled := r.FormValue("enabled") == "on"
	if err := h.App.UpsertChannel(domain.ChannelEmail, enabled, cfg); err != nil {
		http.Error(w, err.Error(), 400); return
	}
	http.Redirect(w, r, "/channels", http.StatusSeeOther)
}

func (h *Handlers) UpsertTelegram(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cfg := map[string]string{
		"botToken": r.FormValue("botToken"),
		"chatID":   r.FormValue("chatID"),
	}
	enabled := r.FormValue("enabled") == "on"
	if err := h.App.UpsertChannel(domain.ChannelTelegram, enabled, cfg); err != nil {
		http.Error(w, err.Error(), 400); return
	}
	http.Redirect(w, r, "/channels", http.StatusSeeOther)
}
