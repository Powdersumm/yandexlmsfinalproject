<<<<<<< HEAD
# Yandex LMS Final Project

Микросервис для асинхронного вычисления математических выражений с JWT-аутентификацией и хранением данных в PostgreSQL.

[![Go](https://img.shields.io/badge/Go-1.19%2B-blue)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14%2B-brightgreen)](https://www.postgresql.org/)

## 🚀 Особенности
- 🔐 Регистрация и аутентификация через JWT
- ➕ Поддержка операций: `+`, `-`, `*`, `/`, скобки
- 📦 Асинхронная обработка задач с очередью
- 📊 История вычислений с фильтрацией по пользователю
- 🐳 Готовая конфигурация для Docker (опционально)

## 🛠 Технологии
- **Backend**: Go (Gin, GORM)
- **База данных**: PostgreSQL
- **Аутентификация**: JWT
- **Очередь задач**: In-memory каналы Go

## ⚙️ Требования
- Go 1.19+
- PostgreSQL 14+
- Переменные окружения (см. [.env.example](.env.example))

## 🚀 Быстрый старт
### Установка PostgreSQL и pgAdmin 4:
Установка pgAdmin 4 обязатлеьно потому что через нее мы будем запускать БД

1. Скачайте и установите PostgreSQL и pgAdmin 4 с официального сайта
2. Создайте базу данных `math_service` и пользователя `math_service` с парол
3. Подключитесь к базе данных через pgAdmin 4:
    -Выбираете серверы
    -Далее PostgreSQL 17
    -Создание базы данных 
    -нужно будет ввести пароль(Обычно это:postgres)
### 1. Клонирование репозитория
```bash
git clone https://github.com/Powdersumm/yandexlmsfinalproject.
```
```bash
cd yandexlmsfinalproject
```
### 2. Настройка окружения
Создайте файл `.env` на основе примера или воспользуйтесь моим файлом который уже создан в корневой папке проекта:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=calculator
JWT_SECRET=your_strong_secret_here
PORT=8080
```

### 3. Запуск PostgreSQL через Docker
```bash
docker-compose up -d
```
### 4. Установка зависимостей
```bash
go mod download
```
### 5. Запуск сервера
```bash
go run ./cmd/main.go
```
### 6. Проверка работоспособности
Для этого нужно будет пройти регистрацию, авторизацию, отправку выражения и получение Get ответа через Postman
Отправьте тестовый запрос для регистрации через Postman:
```bash
POST http://localhost:8080/api/v1/register
Content-Type: application/json

{
  "login": "testuser",
  "password": "testpassword"
}
```
Отправьте тестовый запрос для авторизации через Postman:
```bash
POST http://localhost:8080/api/v1/login
Content-Type: application/json

{
  "login": "testuser",
  "password": "testpassword"
}
```
Отправьте тестовый запрос для отправки математического выражения через Postman:
```bash
POST http://localhost:8080/api/v1/calculate
Authorization: Bearer <your_jwt_token>
Content-Type: application/json

{
  "expression": "(10 + 5) * 2 / 3"
}
```
Отправьте тестовый запрос для Get ответа через Postman:
```bash
GET http://localhost:8080/api/v1/expressions
Authorization: Bearer <your_jwt_token>
```
