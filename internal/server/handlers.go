package server

import (
	"encoding/json"
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
    data := map[string]any{
        "Alerts":      list,
        "Page":        "alerts",
        "ContentTmpl": "alerts_page", // <— tell base which partial to render
    }
    if err := h.tpl.ExecuteTemplate(w, "base", data); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
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
	w.Header().Set("HX-Trigger", "alert-changed")
	_ = h.tpl.ExecuteTemplate(w, "alerts", map[string]any{"Alerts": list})
}

func (h *Handlers) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.App.DeleteAlert(id); err != nil {
		http.Error(w, err.Error(), 400); return
	}
	list, _ := h.App.ListAlerts()
	w.Header().Set("HX-Trigger", "alert-changed")
	_ = h.tpl.ExecuteTemplate(w, "alerts", map[string]any{"Alerts": list})
}

func (h *Handlers) ChannelsPage(w http.ResponseWriter, r *http.Request) {
    chs, _ := h.App.ListChannels()

    // Defaults
    emailEnabled := true
    emailCfg := notif.EmailConfig{}
    tgEnabled := true
    tgCfg := struct {
        BotToken string `json:"botToken"`
        ChatID   string `json:"chatID"`
    }{}

    // Populate from DB rows if present
    for _, ch := range chs {
        switch ch.Kind {
        case domain.ChannelEmail:
            emailEnabled = ch.Enabled
            _ = json.Unmarshal([]byte(ch.Config), &emailCfg)
        case domain.ChannelTelegram:
            tgEnabled = ch.Enabled
            _ = json.Unmarshal([]byte(ch.Config), &tgCfg)
        }
    }

    data := map[string]any{
        "Page":          "channels",
        "EmailEnabled":  emailEnabled,
        "Email":         emailCfg,
        "TGEnabled":     tgEnabled,
        "Telegram":      tgCfg,
        "Saved":         r.URL.Query().Get("saved") == "1",
    }

    if err := h.tpl.ExecuteTemplate(w, "base", data); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
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
        http.Error(w, err.Error(), http.StatusBadRequest); return
    }
    w.Header().Set("HX-Trigger", "channels-saved")
    w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) UpsertTelegram(w http.ResponseWriter, r *http.Request) {
    _ = r.ParseForm()
    cfg := map[string]string{
        "botToken": r.FormValue("botToken"),
        "chatID":   r.FormValue("chatID"),
    }
    enabled := r.FormValue("enabled") == "on"
    if err := h.App.UpsertChannel(domain.ChannelTelegram, enabled, cfg); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest); return
    }
    w.Header().Set("HX-Trigger", "channels-saved")
    w.WriteHeader(http.StatusNoContent)
}
