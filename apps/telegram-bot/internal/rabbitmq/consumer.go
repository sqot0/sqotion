package rabbitmq

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ResultHandler func(msg NotesResultMessage)

func ConsumeResults(ch *amqp.Channel, handler ResultHandler) (<-chan amqp.Delivery, error) {
	deliveries, err := ch.Consume(
		QueueResults,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range deliveries {
			var msg NotesResultMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("failed to unmarshal result message: %v", err)
				continue
			}
			handler(msg)
		}
	}()

	return deliveries, nil
}
