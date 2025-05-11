# Билд приложения
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

# Запуск
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .
COPY .env .

EXPOSE 8080

CMD ["./main"]