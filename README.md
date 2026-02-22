# App

Простое Go-приложение с HTTP-сервером и PostgreSQL.

## Требования

- Go 1.25+
- Docker и Docker Compose (для БД)

## Запуск

1. Инициализация (создаёт `.env` из `example.env` и поднимает PostgreSQL):

   ```bash
   make init
   ```

2. Запуск приложения:

   ```bash
   go run ./cmd/app
   ```

Приложение слушает порт из `HTTP_PORT` (по умолчанию 8080).

## Команды Make

| Команда        | Описание                    |
|----------------|-----------------------------|
| `make init`    | Создать .env и запустить БД |
| `make docker-up`   | Запустить PostgreSQL       |
| `make docker-down` | Остановить PostgreSQL      |
| `make db-shell`    | Подключиться к БД (psql)   |
| `make help`       | Список команд              |

## Конфигурация

Переменные окружения задаются в `.env`. Пример — в `example.env`.
