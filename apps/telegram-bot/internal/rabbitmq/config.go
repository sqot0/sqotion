package rabbitmq

import (
	"fmt"
	"os"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
}

func ConfigFromEnv() Config {
	return Config{
		Host:     envOrDefault("RABBITMQ_HOST", "localhost"),
		Port:     envOrDefault("RABBITMQ_PORT", "5672"),
		User:     envOrDefault("RABBITMQ_USER", "guest"),
		Password: envOrDefault("RABBITMQ_PASSWORD", "guest"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.User, c.Password, c.Host, c.Port)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
