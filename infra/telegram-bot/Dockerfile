FROM golang:1.26.5-alpine AS builder

WORKDIR /build
COPY apps/telegram-bot/ ./apps/telegram-bot/
COPY packages/          ./packages/

WORKDIR /build/apps/telegram-bot
RUN go mod download && CGO_ENABLED=0 go build -o /bin/telegram-bot ./cmd/telegram-bot/

FROM alpine:3.22.5

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/telegram-bot /usr/local/bin/telegram-bot

CMD ["telegram-bot"]
