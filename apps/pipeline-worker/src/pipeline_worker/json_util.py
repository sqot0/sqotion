import json
from typing import Any


def extract_json(text: str | None) -> dict[str, Any] | None:
    if not text:
        return None

    # If it parses cleanly, use it directly.
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass

    # Otherwise try to isolate the JSON object between the first "{" and last "}".
    start = text.find("{")
    end = text.rfind("}")
    if start == -1 or end == -1 or end <= start:
        return None

    candidate = text[start : end + 1]
    try:
        return json.loads(candidate)
    except json.JSONDecodeError as e:
        print(f"{e.msg} at line {e.lineno}, column {e.colno}")
        return None
