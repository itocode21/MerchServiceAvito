# Avito Merch Service

Сервис для обмена монетками и покупки мерча внутри компании Avito.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-336791.svg)
![Redis](https://img.shields.io/badge/Redis-latest-DC382D.svg)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED.svg)

## Описание

Avito Merch Service предоставляет REST API для сотрудников компании:
- Регистрация и аутентификация пользователей (JWT).
- Передача монет между пользователями.
- Покупка мерча за монеты.
- Просмотр информации о монетах, инвентаре и истории транзакций.

Сервис оптимизирован для высокой нагрузки (1000 RPS, p95 < 50 мс, ошибки < 0.01%) и использует PostgreSQL для хранения данных и Redis для кэширования.

## Требования

- **Go**: 1.23+
- **PostgreSQL**: 13+
- **Redis**: latest
- **Docker** и **Docker Compose**
- **Make** (опционально, для удобного запуска команд)

## Установка и запуск

### Локальный запуск

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/itocode21/MerchServiceAvito.git
   cd MerchServiceAvito
2. Создайте файл ```.env ```в корне проекта по примеру ```.env.example```:
    ```env
    DB_HOST=db
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=you_password
    DB_NAME=avito_shop
    DB_SSLMODE=disable
    JWT_SECRET=your_very_secure_secret_key_32_bytes_long
    ```
3. Сборка и запуск:
    ```bash
    make build
    make run
    ```
## Запуск через Docker Compose

1. Запустите сервис с зависимостями:
    ```bash
    make docker-up
    ```
2. Сервис будет доступен на ```http://localhost:8080```.
3. Остановка:
    ```bash
    make docker-down
    ```

## Тестрование

### Юнит-тесты

Покрывают бизнес-логику сервисов (services/item.go, services/transaction.go, services/auth.go). Общее покрытие проекта > 40%.
    ```bash
    make test-unit
    ```
### E2E-тесты

Проверяют ключевые сценарии:

    ```bash
    make test-e2e
    ```
### Нагрузочные тесты

Проверяют производительность (1000 RPS, p95 < 50 мс, ошибки < 0.01%). Используют k6.

    ```bash
    make test-load
    ```
### Результаты:
* RPS: ~2421
* p95: 17.18 мс
* Неожиданные ошибки: 0%

## Эндпоинты
| Метод | Эндпоинт            | Описание                  | Тело запроса (JSON)                       | Заголовки                  |
|-------|---------------------|---------------------------|------------------------------------------|----------------------------|
| POST  | `/api/register`     | Регистрация пользователя  | `{"username": "user1", "password": "12345"}` | `Content-Type: application/json` |
| POST  | `/api/auth`         | Аутентификация (JWT)      | `{"username": "user1", "password": "12345"}` | `Content-Type: application/json` |
| GET   | `/api/info`         | Информация о пользователе | -                                        | `Authorization: Bearer <token>` |
| POST  | `/api/sendCoin`     | Передача монет            | `{"toUser": "user2", "amount": 100}`     | `Authorization: Bearer <token>`<br>`Content-Type: application/json` |
| GET   | `/api/buy/{item}`   | Покупка мерча             | -                                        | `Authorization: Bearer <token>` |

Пример вызова покупки:
    ```bash
    curl -X GET "http://localhost:8080/api/buy/t-shirt" -H "Authorization: Bearer <token>"
    ```

## Структура проекта
    ```text
    MerchServiceAvito/
    ├── cmd/
    │   ├── server/             #Точка входа приложения
    ├── internal/                # Основной код
    │   ├── auth/               # Логика JWT
    │   ├── config/            #Конфигурация
    │   ├── database/        # Миграции
    │   ├── handlers/        # HTTP-обработчики
    │   ├── middleware/  # Middleware (JWT)
    │   ├── models/         # Структуры данных
    │   ├── repositories/  # Работа с базой
    │   └── services/       # Бизнес-логика
    ├── .env.example      # Пример переменных окружения
    ├── docker-compose.yml  # Docker Compose конфигурация
    ├── Dockerfile          # Docker файл
    ├── Makefile            # Команды сборки и тестирования
    └── README.md   # Документация
    ```

## Дополнительно
* Очистка: Удаление бинарников и контейнеров:
    ```bash
    make clean
    ```
* Полная сборка и тестирование:
    ```bash
    make all
    ```

