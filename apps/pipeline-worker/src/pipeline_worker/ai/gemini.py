import json
import mimetypes

import httpx
from google import genai
from google.genai import types

with open("src/pipeline_worker/ai/prompt.txt", "r") as f:
    SYSTEM_INSTRUCTION = f.read()

def call(image_urls: list[str], caption: str | None, api_key: str, model: str) -> str | None:
    """Send images to Gemini and return the full JSON response text."""
    if not api_key:
        print("GEMINI_API_KEY is not set, using mock response", flush=True)
        return json.dumps(
            {
                "path": "inbox/mock-note",
                "frontmatter": {
                    "title": "Mock Note",
                    "description": "Generated with mock mode because GEMINI_API_KEY is not set",
                },
                "markdown": "# Mock Note\n\nThis is a placeholder response.",
            }
        )

    client = genai.Client(api_key=api_key)

    parts: list[types.Part] = []
    for url in image_urls:
        try:
            resp = httpx.get(url, timeout=30)
            resp.raise_for_status()
            content_type = resp.headers.get("content-type", "")
            if content_type.startswith("image/"):
                mime = content_type
            else:
                mime, _ = mimetypes.guess_type(url)
                mime = mime if mime and mime.startswith("image/") else "image/jpeg"
            parts.append(types.Part.from_bytes(data=resp.content, mime_type=mime))
            print(f"  Downloaded image {url} ({mime})", flush=True)
        except Exception as e:
            print(f"  Failed to download {url}: {e}", flush=True)

    if caption:
        parts.append(types.Part.from_text(text=caption))
        print(f"  Added caption: {caption!r}", flush=True)

    if not parts:
        print("No images could be downloaded", flush=True)
        return None

    contents = [types.Content(role="user", parts=parts)]

    try:
        response = client.models.generate_content(
            model=model,
            contents=contents,
            config=types.GenerateContentConfig(
                system_instruction=SYSTEM_INSTRUCTION,
                thinking_config=types.ThinkingConfig(
                    thinking_level="MINIMAL",
                ),
                temperature=0.2,
                response_mime_type="application/json",
                response_schema=types.Schema(
                    type=types.Type.OBJECT,
                    required=["path", "frontmatter", "markdown"],
                    properties={
                        "path": types.Schema(
                            type=types.Type.STRING,
                        ),
                        "frontmatter": types.Schema(
                            type=types.Type.OBJECT,
                            required=["title", "description"],
                            properties={
                                "title": types.Schema(
                                    type=types.Type.STRING,
                                ),
                                "description": types.Schema(
                                    type=types.Type.STRING,
                                ),
                            },
                        ),
                        "markdown": types.Schema(
                            type=types.Type.STRING,
                        ),
                    },
                ),
            ),
        )
        return response.text
    except Exception as e:
        print(f"Gemini API call failed: {e}", flush=True)
        return None
