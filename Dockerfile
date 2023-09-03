FROM golang:1.18.8-alpine3.16 AS build
WORKDIR /go/src/github.com/VATUSA/discord-bot-v3
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
RUN go build -o bin/bot ./cmd/bot/main.go
RUN go build -o bin/web ./cmd/web/main.go

FROM alpine:latest AS bot
WORKDIR /app
COPY --from=build /go/src/github.com/VATUSA/discord-bot-v3/bin/bot ./
COPY config ./config
CMD ["./bot"]

FROM alpine:latest AS web
WORKDIR /app
COPY --from=build /go/src/github.com/VATUSA/discord-bot-v3/bin/web ./
CMD ["./web"]