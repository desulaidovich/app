# App

Go-сервис с gRPC + grpc-gateway и PostgreSQL.

## Быстрый старт

```bash
# 1. Создать .env и поднять PostgreSQL
make init

# 2. Запустить
make run
```

Сервис поднимает два порта:
- `:9090` — gRPC
- `:8080` — HTTP/JSON (grpc-gateway)

## Команды

| Команда           | Описание                              |
|-------------------|---------------------------------------|
| `make init`       | Создать `.env` и запустить БД         |
| `make run`        | Запустить приложение                  |
| `make build`      | Собрать бинарь в `bin/app`            |
| `make proto`      | Сгенерировать код из `.proto` файлов  |
| `make docker-up`  | Запустить PostgreSQL                  |
| `make docker-down`| Остановить PostgreSQL                 |
| `make db-shell`   | Подключиться к БД (psql)              |

## API

| Метод | gRPC | HTTP |
|---|---|---|
| Liveness | `health.v1.HealthService/Health` | `GET /health` |

```bash
# HTTP
curl http://localhost:8080/health

# gRPC (требует APP_DEBUG=true)
grpcurl -plaintext localhost:9090 health.v1.HealthService/Health
```

## Конфигурация

Все переменные задаются в `.env`. Пример — в `example.env`.

| Переменная | По умолчанию | Описание |
|---|---|---|
| `APP_NAME` | — | Имя приложения |
| `APP_ENV` | `development` | Окружение |
| `APP_DEBUG` | `false` | Включает gRPC reflection |
| `HTTP_PORT` | `8080` | Порт grpc-gateway |
| `GRPC_PORT` | `9090` | Порт gRPC |
| `DATABASE_*` | — | Параметры PostgreSQL |
| `LOG_LEVEL` | `debug` | `debug / info / warn / error` |
| `LOG_FORMAT` | `text` | `text / json` |