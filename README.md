# Linkforge

Сервис коротких ссылок на Go: PostgreSQL, срок действия ссылок, rate limiting
по IP, JSON-логи и graceful shutdown.

Стек: net/http, pgx, Goose, Testcontainers.

## Запуск

```bash
make prepare
make docker-up
make migrate-up
```

**Миграции применяются только вручную** через `make migrate-up`.

Команда `make docker-down` также очищает volumes.

## API

```bash
curl -i http://localhost:8080/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com","ttl":"1h"}'
```

Ответ `201` содержит `alias`, `url` и `expires_at`. Без `ttl` ссылка бессрочная.

| Метод  | Путь         | Назначение                                                    |
| ------ | ------------ | ------------------------------------------------------------- |
| `POST` | `/links`     | Создать ссылку с RateLimit                                    |
| `GET`  | `/r/{alias}` | Редирект `302`; для отсутствующей или истёкшей ссылки — `404` |
| `GET`  | `/healthz`   | Состояние HTTP-сервиса                                        |
| `GET`  | `/readyz`    | Доступность PostgreSQL                                        |

## Разработка

| Команда          | Назначение                                     |
| ---------------- | ---------------------------------------------- |
| `make run`       | Запустить API на хосте вместо контейнера       |
| `make test`      | Модульные тесты                                |
| `make test-full` | Все тесты с детектором гонок; требуется Docker |
| `make lint`      | Проверка кода                                  |
| `make format`    | Форматирование                                 |
| `make gen`       | Генерация моков                                |

Интеграционные тесты сами запускают PostgreSQL через Testcontainers.
