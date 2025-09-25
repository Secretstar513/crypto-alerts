# README.md

# Crypto Alerts (Go)

Personal crypto **price-threshold alert** service with a sleek web UI, **Binance** price streaming, and **pluggable notification channels** (Log, Email, Telegram). Single-user by design, easy to extend.

---

## ✨ Features

- Create / enable / disable / delete **alerts**
  - Format: **Symbol** (e.g., `BTCUSDT`) + **Threshold** + **Direction** (**UP**/*crossing upward* or **DOWN**/*crossing downward*)
- **Live prices** via Binance **WebSocket**, with **HTTP fallback** if WS fails
- **Notification channels** via a clean interface:
  - ✅ Log (always on)
  - ✅ Email (SMTP) — MailHog ready for local dev
  - ✅ Telegram bot
  - ➕ Add your own by implementing one interface
- **Validation** (uppercase symbols, positive threshold, valid direction)
- **Nice dark UI** (HTMX + minimal CSS), with toasts, confirm dialogs, and responsive layout
- **SQLite** persistence (pure-Go driver; **no CGO**)
- Single binary, zero external deps (MailHog optional)

---

## 🧱 Tech Stack

- **Go** 1.22+
- **net/http**, **chi** router
- **HTMX** (no framework on client)
- **GORM** + **glebarez/sqlite** (pure Go SQLite)
- **nhooyr/websocket** (Binance stream)
- **zerolog** (logging)

---

## 📦 Project Structure

```
crypto-alerts/
  cmd/
    server/
      main.go
  internal/
    app/           # orchestration (engine, alert checks, notifier fanout)
    config/        # env/.env loader
    db/            # SQLite open
    domain/        # models + validators
    notif/         # Notifier interface + log/email/telegram
    price/         # WS stream + HTTP fallback + symbol stream router
    rules/         # crossing rule
    server/        # handlers, routes, template loader
  web/
    static/        # style.css, htmx.min.js
    templates/     # base + pages + partials
  .env.example
  README.md
  go.mod
```

---

## 🚀 Quick Start

### 0) Prereqs

- Go 1.22+
- (Optional for local email) Docker

### 1) Clone & config

```bash
git clone https://github.com/Secretstar513/crypto-alerts.git
cd crypto-alerts
cp .env.example .env
```

`.env` defaults:

```ini
ADDR=:8080
DB_PATH=alerts.db

# Optional email (use MailHog for local dev)
SMTP_HOST=localhost
SMTP_PORT=1025
EMAIL_FROM=alerts@example.com
EMAIL_TO=you@example.com

# Optional telegram
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
```

> You can also bind to localhost only:
> `ADDR=127.0.0.1:8080`

### 2) Run

```bash
go run ./cmd/server
# server logs:
# Config loaded: addr=:8080 db=alerts.db
# {"level":"info","addr":":8080","message":"server listening"}
```

Open: **http://localhost:8080**

---

## 🖥️ Using the App

### Alerts Page

1. **Create Alert**  
   Example:
   - Symbol: `BTCUSDT`
   - Threshold: `65000`
   - Direction: `UP` (fires when price crosses upward through 65000)

2. **Tips to test quickly**  
   Find current price:
   ```bash
   curl "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"
   ```
   Create **two** alerts:
   - `UP` with **threshold = current − 20**
   - `DOWN` with **threshold = current + 20**

   Normal price wiggles will cross one side soon.  
   When it fires you’ll see a log line like:
   ```
   {"level":"info","notifier":"log","symbol":"BTCUSDT","price":..., "threshold":..., "direction":"UP","message":"ALERT"}
   ```

3. **Manage**  
   - **Enable/Disable** toggles alert active state
   - **Delete** asks for confirm (`hx-confirm`) and then removes the alert

### Channels Page

- **Email**  
  For local testing run MailHog:
  ```bash
  docker run --rm -p 1025:1025 -p 8025:8025 mailhog/mailhog
  ```
  Then set in the Email form:  
  `Host=localhost`, `Port=1025`, `From`, `To`.  
  Open **http://localhost:8025** to see messages when alerts fire.

- **Telegram**  
  - Create a bot with **@BotFather** → get **Bot Token**  
  - Start a chat with your bot, then find your `chat_id`:
    ```bash
    curl "https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates"
    ```
    Use the `chat.id` from the JSON.
  - Fill **Bot Token** and **Chat ID** in the form.

> The UI saves channel configs to the DB. If you implemented the optional `App.Reload()` hook, notifiers pick up saved config immediately without restarting.

---

## 🧠 How Crossing Works

For each symbol with active alerts, the engine keeps a last price. On every update:

- **UP** fires if `prev < threshold` and `current >= threshold`
- **DOWN** fires if `prev > threshold` and `current <= threshold`

We ignore the first tick per symbol (need a baseline).  
Stream source is Binance WS; if it fails, the app polls HTTP ticker every ~10 seconds.

---

## 🔐 Validation Rules

- **Symbol** must be **uppercase** (e.g. `BTCUSDT`, `ETHUSDT`)
- **Threshold** must be `> 0`
- **Direction** ∈ {`UP`, `DOWN`}

Invalid input yields a `400` on creation; the UI shows an error toast.

---

## 🔧 Configuration Reference

Environment (via `.env` or real env vars):

| Var                 | Default         | Description                          |
|---------------------|-----------------|--------------------------------------|
| `ADDR`              | `:8080`         | HTTP listen address                  |
| `DB_PATH`           | `alerts.db`     | SQLite file path                     |
| `SMTP_HOST`         |                 | SMTP host (e.g., `localhost`)        |
| `SMTP_PORT`         |                 | SMTP port (e.g., `1025`)             |
| `SMTP_USER`/`PASS`  |                 | SMTP auth (if needed)                |
| `EMAIL_FROM`        |                 | From address                         |
| `EMAIL_TO`          |                 | To address                           |
| `TELEGRAM_BOT_TOKEN`|                 | Bot token from BotFather             |
| `TELEGRAM_CHAT_ID`  |                 | Your chat or group ID                |

> If Email/Telegram aren’t set, the corresponding notifier simply no-ops.

---

## 🧪 Manual Test Checklist

- [ ] Create two alerts around current `BTCUSDT` price → one should fire within a few minutes.  
- [ ] Toggle alert **Enable/Disable** and verify no alerts fire while disabled.  
- [ ] **Delete** → confirm dialog → row disappears only on “OK”.  
- [ ] Channels page **Save** → toast appears, config persists across page reloads.  
- [ ] With MailHog running, **Email** arrives on alert.  
- [ ] With Telegram configured, **message** arrives on alert.

---

## 📡 API (Internal)

- `POST /alerts` → create (HTMX partial response)
- `POST /alerts/{id}/toggle` → enable/disable (HTMX)
- `POST /alerts/{id}/delete` → delete (HTMX, confirm via `hx-confirm`)
- `GET /channels` → channels page
- `POST /channels/email` → save email config (returns `204`, triggers `channels-saved`)
- `POST /channels/telegram` → save tg config (returns `204`, triggers `channels-saved`)

---

## 🛠️ Dev Notes

- Templates are split into:
  - `base.tmpl.html` (layout + toasts + page switch)
  - `index.tmpl.html` (`alerts_page`)
  - `alerts.tmpl.html` (alerts table partial, returned for HTMX swaps **including wrapper** with `id="alerts-list"`)
  - `channels.tmpl.html` (`channels_page`)
- HTMX is served locally at `/static/htmx.min.js` to avoid third-party script quirks.

---

## 📄 License

MIT (or your choice). See `LICENSE` if provided.

