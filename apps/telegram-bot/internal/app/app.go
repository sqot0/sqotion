package app

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	tb "gopkg.in/telebot.v4"
	"sqotion/apps/telegram-bot/internal/bot"
	"sqotion/apps/telegram-bot/internal/config"
	"sqotion/apps/telegram-bot/internal/rabbitmq"
)

func Run() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Load()

	botInstance, err := tb.NewBot(tb.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	conn, err := amqp.Dial(cfg.RabbitMQ.DSN())
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}

	if err := rabbitmq.DeclareTopology(ch); err != nil {
		log.Fatalf("failed to declare topology: %v", err)
	}

	handler := bot.NewHandler(ch)
	handler.Register(botInstance)

	if _, err := rabbitmq.ConsumeResults(ch, func(msg rabbitmq.NotesResultMessage) {
		handler.HandleResult(msg)
		original := &tb.Message{
			ID:   msg.MessageID,
			Chat: &tb.Chat{ID: msg.ChatID},
		}

		var replyText string

		if msg.Status == "completed" {
		 replyText = "✅ Note has been added successfully!\n" + msg.Text
		} else {
			replyText = "Failed to add note"
		}

		if _, err := botInstance.Reply(original, replyText); err != nil {
			log.Printf("failed to reply to message %d: %v", msg.MessageID, err)
		}
	}); err != nil {
		log.Fatalf("failed to start result consumer: %v", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	go botInstance.Start()
	log.Println("telegram-bot is running...")

	<-ctx.Done()
	log.Println("shutting down...")

	botInstance.Stop()
	ch.Close()
	conn.Close()
}
