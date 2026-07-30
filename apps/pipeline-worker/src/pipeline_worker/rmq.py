import json

import pika


def declare_topology(ch):
    """Declare the exchange, queues, and bindings."""
    ch.exchange_declare(exchange="sqotion", exchange_type="topic", durable=True)
    ch.queue_declare(queue="notes.images", durable=True)
    ch.queue_bind(queue="notes.images", exchange="sqotion", routing_key="notes.images")
    ch.queue_declare(queue="notes.results", durable=True)
    ch.queue_bind(
        queue="notes.results", exchange="sqotion", routing_key="notes.results"
    )


def publish_result(
    ch, batch_id: str, chat_id: int, message_id: int, text: str | None
):
    """Publish a completion result back to the notes.results queue."""
    status = "completed" if text else "failed"
    msg = {
        "batch_id": batch_id,
        "chat_id": chat_id,
        "message_id": message_id,
        "status": status,
        "text": text or "",
    }
    ch.basic_publish(
        exchange="sqotion",
        routing_key="notes.results",
        body=json.dumps(msg),
        properties=pika.BasicProperties(
            content_type="application/json",
            delivery_mode=2,  # persistent
        ),
    )
    print(f"Published result: batch={batch_id} status={status}")
