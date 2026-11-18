# SyncBoard

Синхронизированная канбан-доска: Go REST API + PostgreSQL, React/TypeScript frontend и WebSocket-события для обновления открытых досок.

## Требования

- Go 1.25.1 (версия из `syncboard-backend/go.mod`)
- Node.js 20+
- Docker и Docker Compose
- golang-migrate CLI для миграций

## Запуск backend

Откройте отдельный терминал из корня проекта:

```bash
cd syncboard-backend
cp .env.example .env
docker compose up -d postgres
migrate -path migrations -database "postgres://postgres:postgres123@localhost:5432/syncboard?sslmode=disable" up
go run ./cmd/api
```

При первом запуске Go автоматически загрузит зависимости, указанные в `go.mod`. Их контрольные суммы должны быть записаны в `go.sum`; в этой версии проекта контрольные суммы `github.com/gorilla/websocket v1.5.3` уже добавлены.

### Если Go сообщает `missing go.sum entry`

Эта ошибка означает, что для зависимости не записана контрольная сумма в `go.sum`. Не нужно выполнять `go get syncboard/internal/realtime`: `syncboard/internal/realtime` — это пакет самого проекта, а не внешний модуль.

Выполните команды **из папки `syncboard-backend`**:

```bash
go mod tidy
go mod download
go run ./cmd/api
```

`go mod tidy` проверит импорты и приведёт `go.mod`/`go.sum` в соответствие с используемыми пакетами. Если команда изменит эти файлы, сохраните изменения в Git:

```bash
git add go.mod go.sum
git commit -m "fix: update Go module checksums"
```

Если загрузка зависимостей завершается сетевой ошибкой, проверьте доступ к `proxy.golang.org` и повторите команды после восстановления соединения.

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
- создание и просмотр досок через интерфейс сайта;
- создание колонок и карточек через встроенные модальные формы (без браузерных `prompt`/`confirm`);
- перетаскивание карточек между колонками и внутри колонок;
- перетаскивание колонок по горизонтали с сохранением порядка в PostgreSQL;
- приглашения участников по email, входящие приглашения, принятие и отклонение;
- просмотр списка участников доски; отправлять приглашения могут администраторы;
- проверка членства пользователя в доске для REST и WebSocket;
- WebSocket-события `column.created`, `column.reordered`, `card.created`, `card.moved`, `card.deleted`, `invitation.created`, `member.joined`;
- обновление открытой доски при получении события;
- транзакционное создание доски и изменение порядка колонок.

Для новой схемы примените все миграции, включая `000005_add_invitations_and_column_order`.

Приглашение пока не отправляет электронное письмо: запись создаётся в базе, а приглашённый пользователь видит её в блоке «Входящие приглашения» после входа в аккаунт с указанным email. Для реальной email-рассылки потребуется подключить SMTP или почтовый сервис.

## Примечания по безопасности

- В WebSocket JWT передаётся в query-параметре, поскольку стандартный браузерный WebSocket API не позволяет передать `Authorization` header. В production предпочтительнее короткоживущий одноразовый ticket или cookie с `HttpOnly`, `Secure`, `SameSite`.
- Замените `JWT_SECRET` на криптографически случайный секрет. Не используйте `.env` в репозитории.
- Текущая WebSocket-рассылка работает в рамках одного процесса backend. Для нескольких экземпляров понадобится Redis Pub/Sub или другой брокер событий.
