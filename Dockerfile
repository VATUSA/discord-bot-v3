FROM golang:1.18.8-alpine3.16 AS build
WORKDIR /go/src/github.com/VATUSA/discord-bot-v3
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
RUN go build -o bin/bot ./cmd/bot/main.go
RUN go build -o bin/web ./cmd/web/main.go

FROM alpine:latest AS app
WORKDIR /app
COPY --from=build /go/src/github.com/VATUSA/discord-bot-v3/bin/* ./