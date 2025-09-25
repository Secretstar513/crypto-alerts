package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/Secretstar513/crypto-alerts/internal/config"
	"github.com/Secretstar513/crypto-alerts/internal/db"
	"github.com/Secretstar513/crypto-alerts/internal/domain"
	"github.com/Secretstar513/crypto-alerts/internal/notif"
	"github.com/Secretstar513/crypto-alerts/internal/price"
	"github.com/Secretstar513/crypto-alerts/internal/rules"
)

type App struct {
	Cfg        *config.Config
	DB         *gorm.DB
	Router     *price.Router
	Notifiers  []notif.Notifier
	cancel     context.CancelFunc
}

func New(cfg *config.Config) *App {
	d := db.OpenSQLite(cfg.DBPath)
	// migrations
	if err := d.AutoMigrate(&domain.Alert{}, &domain.Channel{}, &domain.LastPrice{}); err != nil {
		panic(err)
	}

	// built-in channels: Log is always present (enabled true by default)
	notifs := []notif.Notifier{
		notif.NewLog(true),
		notif.NewEmail(notif.EmailConfig{
			Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass,
			From: cfg.EmailFrom, To: cfg.EmailTo,
		}, true), // enable; it will no-op if not configured
		notif.NewTelegram(cfg.TelegramBotToken, cfg.TelegramChatID, true),
	}

	return &App{
		Cfg:       cfg,
		DB:        d,
		Router:    price.NewRouter(),
		Notifiers: notifs,
	}
}

func (a *App) Start(ctx context.Context) {
	ctx, a.cancel = context.WithCancel(ctx)
	go a.runEngine(ctx)
}

func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	a.Router.StopAll()
}

func (a *App) runEngine(ctx context.Context) {
	// Subscribe per symbol for all enabled alerts and evaluate crossings.
	// Reconcile every 5s to catch new/changed alerts without restart.
	tk := time.NewTicker(5 * time.Second)
	defer tk.Stop()

	type subInfo struct {
		sub price.Subscriber
	}
	subs := map[string]subInfo{} // key: symbol

	for {
		// reconcile
		var alerts []domain.Alert
		if err := a.DB.Where("enabled = ?", true).Find(&alerts).Error; err == nil {
			need := map[string]struct{}{}
			for _, al := range alerts {
				need[al.Symbol] = struct{}{}
			}
			for sym := range need {
				if _, ok := subs[sym]; !ok {
					subs[sym] = subInfo{sub: a.Router.Subscribe(ctx, sym)}
				}
			}
		}

		// pump updates non-blockingly
		for sym, si := range subs {
			select {
			case upd := <-si.sub:
				a.handlePriceUpdate(ctx, upd.Symbol, upd.Price)
			default:
				_ = sym
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-tk.C:
		}
	}
}

func (a *App) handlePriceUpdate(ctx context.Context, symbol string, priceVal float64) {
	var lp domain.LastPrice
	if err := a.DB.First(&lp, "symbol = ?", symbol).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		lp = domain.LastPrice{Symbol: symbol, Price: priceVal, UpdatedAt: time.Now()}
		a.DB.Save(&lp)
		return
	}
	prev := lp.Price
	lp.Price = priceVal
	lp.UpdatedAt = time.Now()
	a.DB.Save(&lp)

	// Check alerts for this symbol
	var alerts []domain.Alert
	if err := a.DB.Where("enabled = ? AND symbol = ?", true, symbol).Find(&alerts).Error; err != nil {
		return
	}

	for _, al := range alerts {
		if rules.Crosses(prev, priceVal, &al) {
			a.fire(ctx, al, priceVal)
		}
	}
}

func (a *App) fire(ctx context.Context, al domain.Alert, priceVal float64) {
	ev := notif.Event{
		Symbol: al.Symbol, Price: priceVal, Threshold: al.Threshold, Direction: string(al.Direction),
	}
	for _, n := range a.Notifiers {
		if n.Enabled() {
			if err := n.Notify(ctx, ev); err != nil {
				log.Error().Err(err).Str("notifier", n.Name()).Msg("notify failed")
			}
		}
	}
}

func (a *App) CreateAlert(symbol string, thr float64, dir domain.Direction) (domain.Alert, error) {
	al := domain.Alert{
		ID:        nuid.New(),
		Symbol:    symbol,
		Threshold: thr,
		Direction: dir,
		Enabled:   true,
	}
	if err := domain.ValidateAlert(&al); err != nil {
		return al, err
	}
	return al, a.DB.Create(&al).Error
}

func (a *App) ToggleAlert(id string, enabled bool) error {
	return a.DB.Model(&domain.Alert{}).Where("id = ?", id).Update("enabled", enabled).Error
}

func (a *App) DeleteAlert(id string) error {
	return a.DB.Delete(&domain.Alert{}, "id = ?", id).Error
}

func (a *App) ListAlerts() ([]domain.Alert, error) {
	var list []domain.Alert
	return list, a.DB.Order("created_at desc").Find(&list).Error
}

func (a *App) UpsertChannel(kind domain.ChannelKind, enabled bool, cfg any) error {
	js, _ := json.Marshal(cfg)
	var ch domain.Channel
	res := a.DB.First(&ch, "kind = ?", kind)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		ch = domain.Channel{
			ID: nuid.New(), Kind: kind, Enabled: enabled, Config: string(js),
		}
		return a.DB.Create(&ch).Error
	}
	ch.Enabled = enabled
	ch.Config = string(js)
	return a.DB.Save(&ch).Error
}

func (a *App) ListChannels() ([]domain.Channel, error) {
	var out []domain.Channel
	return out, a.DB.Order("kind asc").Find(&out).Error
}
