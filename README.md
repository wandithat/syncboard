# SyncBoard

Синхронизированная канбан-доска: Go REST API + PostgreSQL, React/TypeScript frontend и WebSocket-события для обновления открытых досок.

## Требования

- Go 1.25.1 (версия из `syncboard-backend/go.mod`)
- Node.js 20+
- Docker и Docker Compose
- golang-migrate CLI для миграций

## Запуск backend

```bash
cd syncboard-backend
cp .env.example .env
docker compose up -d postgres
migrate -path migrations -database "postgres://postgres:postgres123@localhost:5432/syncboard?sslmode=disable" up
go run ./cmd/api
```

Backend слушает `http://localhost:8080`; health check: `GET /health`.

## Запуск frontend

В другом терминале:

```bash
cd syncboard-frontend
npm install
npm run dev
```

Vite по умолчанию доступен на `http://localhost:5173`. При необходимости задайте `VITE_API_URL` и `VITE_WS_URL` (по умолчанию `http://localhost:8080/api` и `ws://localhost:8080`).

## Реализованные возможности

- регистрация и вход по JWT;
- создание и просмотр досок;
- создание колонок и карточек;
- перемещение карточек с сохранением порядка;
- проверка членства пользователя в доске для REST и WebSocket;
- WebSocket-канал `/ws/boards/{id}?token=<JWT>` с событиями `column.created`, `card.created`, `card.moved`, `card.deleted`;
- обновление открытой доски при получении события;
- транзакционное создание доски и управление порядком карточек.

## Примечания по безопасности

- В WebSocket JWT передаётся в query-параметре, поскольку стандартный браузерный WebSocket API не позволяет передать `Authorization` header. В production предпочтительнее короткоживущий одноразовый ticket или cookie с `HttpOnly`, `Secure`, `SameSite`.
- Замените `JWT_SECRET` на криптографически случайный секрет. Не используйте `.env` в репозитории.
- Текущая WebSocket-рассылка работает в рамках одного процесса backend. Для нескольких экземпляров понадобится Redis Pub/Sub или другой брокер событий.
