# Этап сборки
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/main.go

# Этап запуска
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
# Если вы хотите использовать .env файл:
COPY .env /root/.env
EXPOSE 8080
# Если .env не подгружается автоматически, можно использовать
CMD ["sh", "-c", "source /root/.env && ./app"]
