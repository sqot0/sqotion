package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		Exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return err
	}

	qNotes, err := ch.QueueDeclare(
		QueueNotes,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	if err := ch.QueueBind(
		qNotes.Name,
		RoutingKeyNote,
		Exchange,
		false,
		nil,
	); err != nil {
		return err
	}

	qResults, err := ch.QueueDeclare(
		QueueResults,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	return ch.QueueBind(
		qResults.Name,
		RoutingKeyResult,
		Exchange,
		false,
		nil,
	)
}

func PublishImage(ch *amqp.Channel, ctx context.Context, msg ImageMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal message: %v", err)
		return err
	}

	return ch.PublishWithContext(ctx,
		Exchange,
		RoutingKeyNote,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func PublishImageBatch(ch *amqp.Channel, ctx context.Context, msg ImageBatchMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal batch message: %v", err)
		return err
	}

	return ch.PublishWithContext(ctx,
		Exchange,
		RoutingKeyNote,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func PublishNoteResult(ch *amqp.Channel, ctx context.Context, msg NotesResultMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal result message: %v", err)
		return err
	}

	return ch.PublishWithContext(ctx,
		Exchange,
		RoutingKeyResult,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}
