import json

from pika import BlockingConnection

from pipeline_worker.config import load
from pipeline_worker.gemini import call as call_gemini
from pipeline_worker.rmq import declare_topology, publish_result
from pipeline_worker.storage import save_note
from pipeline_worker.telegram import fetch_file_url


def process_batch(
    ch, method, properties, body, bot_token, api_key, model, s3_bucket
):
    print(">>> process_batch called", flush=True)
    data = json.loads(body)

    batch_id = data.get("batch_id")
    chat_id = data.get("chat_id")
    message_id = data.get("message_id")
    file_ids: list[str] = data.get("file_ids", [])
    caption: str | None = data.get("caption")

    print(
        f"Processing batch: batch_id={batch_id} chat_id={chat_id} "
        f"file_ids={file_ids} caption={caption!r}",
        flush=True,
    )

    image_urls = []
    for fid in file_ids:
        url = fetch_file_url(fid, bot_token)
        if url:
            image_urls.append(url)
            print(f"  Resolved file_id={fid} -> {url}", flush=True)
        else:
            print(f"  Failed to resolve file_id={fid}", flush=True)

    if not image_urls:
        print("No images could be resolved, marking as failed", flush=True)
        publish_result(ch, batch_id, chat_id, message_id, None)
        return

    print(f"Calling Gemini with {len(image_urls)} image(s)...", flush=True)
    text = call_gemini(image_urls, caption, api_key, model)

    if text:
        print(f"Gemini response:\n{text}", flush=True)
        try:
            result = json.loads(text)
            rel_path = result.get("path", "unknown")
            frontmatter = result.get("frontmatter", {})
            markdown_body = result.get("markdown", "")

            save_note(
                s3_bucket, rel_path, frontmatter, markdown_body, image_urls, file_ids
            )

            title = frontmatter.get("title", "") or rel_path.split("/")[-1]
            formatted = f"Title: {title}\nPath: {rel_path}"
            publish_result(ch, batch_id, chat_id, message_id, formatted)
        except json.JSONDecodeError:
            print("Failed to parse Gemini response as JSON", flush=True)
            publish_result(ch, batch_id, chat_id, message_id, text)
    else:
        print("Gemini returned no response", flush=True)
        publish_result(ch, batch_id, chat_id, message_id, None)


def main():
    connection_params, api_key, model, bot_token, s3_bucket = load()

    print(
        f"Starting pipeline worker (api_key={'set' if api_key else 'not set'}, "
        f"bot_token={'set' if bot_token else 'not set'}, "
        f"s3_bucket={s3_bucket})",
        flush=True,
    )

    with BlockingConnection(connection_params) as conn, conn.channel() as ch:
        declare_topology(ch)

        ch.basic_consume(
            queue="notes.images",
            on_message_callback=lambda ch, method, properties, body: process_batch(
                ch, method, properties, body, bot_token, api_key, model, s3_bucket
            ),
            auto_ack=True,
        )
        print("Pipeline worker is running. Waiting for messages...", flush=True)
        ch.start_consuming()
