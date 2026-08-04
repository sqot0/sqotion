import time

import httpx

_MAX_ATTEMPTS = 3
_RETRY_DELAY_SECONDS = 2


def trigger_deploy(url: str, timeout: float = 60.0) -> bool:
    url = (url or "").strip()
    if not url:
        return False

    last_error: Exception | None = None
    for attempt in range(1, _MAX_ATTEMPTS + 1):
        try:
            resp = httpx.post(url, timeout=timeout)
            resp.raise_for_status()
            print(f"  Deploy hook triggered: {url} (attempt {attempt})", flush=True)
            return True
        except httpx.HTTPError as e:
            last_error = e
            print(
                f"  Deploy hook attempt {attempt}/{_MAX_ATTEMPTS} failed: {e}",
                flush=True,
            )
            if attempt < _MAX_ATTEMPTS:
                time.sleep(_RETRY_DELAY_SECONDS)

    print(
        f"  Deploy hook failed after {_MAX_ATTEMPTS} attempts: {last_error}", flush=True
    )
    return False
