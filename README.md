# Avito Merch Service

Сервис для обмена монетками и покупки мерча внутри компании Avito, разработанный itocode21.

## Описание
Сервис предоставляет API для:
- Регистрации и аутентификации пользователей (JWT).
- Передачи монет между пользователями.
- Покупки мерча за монеты.
- Просмотра информации о монетах, инвентаре и истории транзакций.

## Требования
- Go 1.23+
- PostgreSQL 13+
- Docker и Docker Compose
- Make (опционально, для использования Makefile)

## Установка и запуск

### Локальный запуск
1. Склонируйте репозиторий:
   ```bash
   git clone <github.com/itocode21/MerchServiceAvito>
   cd MerchServiceAvito
   ```
2. Создайте файл ```.env``` в корне проекта по примеру ```.env.example```:
    ```text
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=your_password
    DB_NAME=avito_shop
    DB_SSLMODE=disable
    JWT_SECRET=your_secret_key
    ```
3. Убедитесь, что PostgreSQL запущена и база ```avito_shop``` создана.
   
4. Запустите приложение:
    ```bash
    go run ./cmd/server/main.go
    ```
5. Запуск через Docker Compose:
    ```bash
    docker-compose up --build
    ```
Сервис будет доступен на ```localhost:8080```

## Тестирование
* Юнит тесты(покрытие >40%):
    ```bash
    go test ./internal/services -v -cover
    ```
* E2E-тесты(покупка мерча и передача монет):
    ```bash
    go test -v .
    ```

## Эндпоинты

* ```POST /api/register``` - Регистрация пользователя:
    ```json
    {"username": "user1", "password": "12345"}
    ```

* ```POST /api/auth``` - Аутентификация(возвращает JWT):
    ```json
    {"username": "user1", "password": "12345"}
    ```

* ```GET /api/info``` - Информация о пользователе (требует JWT):
    ```text
    Authorization: Bearer <token>
    ```

* ```POST /api/sendCoi``` - Передача монет (требует JWT):
    ```json
    {"toUser": "user2", "amount": 100}
    ```

* ```GET /api/buy/{item}``` - Покупка мерча(требует JWT):
    ```text
    GET /api/buy/t-shirt
    ```
