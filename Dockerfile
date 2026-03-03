# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/server ./cmd/server

FROM gcr.io/distroless/base-debian12 AS runner
WORKDIR /app
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/.env /app/.env

EXPOSE 3000
ENTRYPOINT ["/app/server"]
