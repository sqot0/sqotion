FROM python:3.13-slim AS builder

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

WORKDIR /app
COPY apps/pipeline-worker/ ./
RUN uv sync --frozen --no-dev --no-editable

FROM python:3.13-slim

COPY --from=builder /app/.venv /app/.venv
COPY --from=builder /app/src /app/src
COPY --from=builder /app/pyproject.toml /app/

WORKDIR /app
ENV PATH="/app/.venv/bin:$PATH"

CMD ["pipeline-worker"]
