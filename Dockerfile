FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api-proxy ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-proxy .
COPY config.yml .

EXPOSE 8080

CMD ["./api-proxy"]
