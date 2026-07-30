package config

import (
	"os"
	"sqotion/apps/telegram-bot/internal/rabbitmq"
)

type App struct {
	TelegramToken string
	RabbitMQ      rabbitmq.Config
}

func Load() App {
	return App{
		TelegramToken: requireEnv("TELEGRAM_BOT_TOKEN"),
		RabbitMQ:      rabbitmq.ConfigFromEnv(),
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable " + key + " is not set")
	}
	return v
}
