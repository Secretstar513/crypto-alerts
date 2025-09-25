package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr           string
	DBPath         string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	EmailFrom      string
	EmailTo        string
	TelegramBotToken string
	TelegramChatID   string
}

func Load() *Config {
	_ = godotenv.Load()

	c := &Config{
		Addr:     get("ADDR", ":8080"),
		DBPath:   get("DB_PATH", "alerts.db"),
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: os.Getenv("SMTP_PORT"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		EmailFrom: os.Getenv("EMAIL_FROM"),
		EmailTo:   os.Getenv("EMAIL_TO"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}

	log.Printf("Config loaded: addr=%s db=%s", c.Addr, c.DBPath)
	return c
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
