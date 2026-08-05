package config

import (
	"os"
	"sqotion/apps/telegram-bot/internal/rabbitmq"
)

type App struct {
	TelegramToken string
	PublicUiDeployHookUrl string
	RabbitMQ      rabbitmq.Config
}

func Load() App {
	return App{
		TelegramToken: requireEnv("TELEGRAM_BOT_TOKEN"),
		PublicUiDeployHookUrl: os.Getenv("PUBLIC_UI_DEPLOY_HOOK_URL"),
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
