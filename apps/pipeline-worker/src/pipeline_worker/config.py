import os

import pika
from pika import ConnectionParameters


def load() -> tuple[ConnectionParameters, str, str, str, str, str, str]:
    """Load all configuration from environment variables."""
    host = os.getenv("RABBITMQ_HOST", "localhost")
    port = int(os.getenv("RABBITMQ_PORT", "5672"))
    user = os.getenv("RABBITMQ_USER", "guest")
    password = os.getenv("RABBITMQ_PASSWORD", "guest")
    gemini_api_key = os.getenv("GEMINI_API_KEY", "")
    gemini_model = os.getenv("GEMINI_MODEL", "gemini-3.1-flash-lite")
    telegram_bot_token = os.getenv("TELEGRAM_BOT_TOKEN", "")
    s3_bucket = os.getenv("S3_BUCKET", "sqotion-notes")
    s3_prefix = os.getenv("S3_PREFIX", "")
    deploy_hook_url = os.getenv("PUBLIC_UI_DEPLOY_HOOK_URL", "")

    connection_params = ConnectionParameters(
        host=host, port=port, credentials=pika.PlainCredentials(user, password)
    )
    return connection_params, gemini_api_key, gemini_model, telegram_bot_token, s3_bucket, s3_prefix, deploy_hook_url
