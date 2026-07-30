import httpx


def fetch_file_url(file_id: str, bot_token: str) -> str | None:
    """Resolve a Telegram file_id to a direct HTTPS download URL."""
    if not bot_token:
        print("TELEGRAM_BOT_TOKEN is empty, cannot fetch file URL", flush=True)
        return None

    url = f"https://api.telegram.org/bot{bot_token}/getFile"
    try:
        with httpx.Client() as client:
            resp = client.post(url, json={"file_id": file_id}, timeout=10)
            data = resp.json()
            if not data.get("ok"):
                print(
                    f"Failed to get file path for {file_id}: {data}",
                    flush=True,
                )
                return None
            file_path = data["result"]["file_path"]
            return f"https://api.telegram.org/file/bot{bot_token}/{file_path}"
    except Exception as e:
        print(f"Error fetching file URL for {file_id}: {e}", flush=True)
        return None
