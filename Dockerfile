# --platform=$BUILDPLATFORM keeps the build stage on the runner's native arch and
# lets Go cross-compile to $TARGETARCH instead of emulating the whole toolchain
# under QEMU. For a pure-Go binary this is the same output at a fraction of the
# build time, so multi-arch costs us almost nothing here.
FROM --platform=$BUILDPLATFORM golang:1.18.8-alpine3.16 AS build
WORKDIR /go/src/github.com/VATUSA/discord-bot-v3
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
ARG TARGETOS
ARG TARGETARCH
# CGO_ENABLED=0 is required for cross-compilation without a target C toolchain,
# and is safe here: the binary is pure Go.
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/bot ./cmd/bot/main.go

FROM alpine:latest AS app
WORKDIR /app
EXPOSE 9002
COPY --from=build /go/src/github.com/VATUSA/discord-bot-v3/bin/* ./
