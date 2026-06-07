# Somon CRM — Backend (Go + Gin)

Backend-и нави Somon CRM, ки бо Go + Gin + PostgreSQL + Redis + JWT навишта шудааст,
бо интегратсияи Telegram bot барои огоҳсозӣ.

> Бекенди тоҷикӣ ва русиро дастгирӣ мекунад (i18n).

---

## 🛠 Tech stack

| Категория          | Технология                          |
|--------------------|-------------------------------------|
| Language           | Go 1.23+                            |
| HTTP framework     | Gin                                 |
| ORM                | GORM v2 + PostgreSQL                |
| Cache / sessions   | Redis (token blacklist + tg codes)  |
| Auth               | JWT (HS256) + bcrypt                |
| Migrations         | golang-migrate (SQL files)          |
| Telegram           | go-telegram-bot-api/v5              |
| Excel export       | excelize/v2                         |
| i18n               | tg, ru                              |
| Container          | Docker + docker-compose + Nginx     |

---

## 📁 Сохтор

```
backend/
├── cmd/
│   ├── server/        # main HTTP сервер
│   └── migrate/       # CLI барои миграсияҳо
├── internal/
│   ├── config/        # .env / viper
│   ├── database/      # postgres + redis
│   ├── i18n/          # бундлҳои tg/ru
│   ├── middleware/    # JWT, CORS, locale
│   ├── models/        # GORM моделҳо
│   ├── service/       # auth, telegram
│   ├── handler/       # HTTP handler-ҳо (домен-домен)
│   ├── router/        # сабти ҳамаи route-ҳо
│   └── utils/         # upload, excel, slug, response
├── migrations/        # 0001_init.up/down.sql
├── uploads/           # файлҳои корбар (Docker volume)
├── Dockerfile
├── docker-compose.yml
├── nginx.conf
├── go.mod
└── README.md
```

---

## ⚡ Шурӯъ кардан

### 1. Талаботҳо

- Go 1.23+
- Docker + docker-compose (тавсия дода мешавад)

### 2. Бо Docker (осонтарин)

```bash
cp .env.example .env
# .env-ро таҳрир намоед — ҳадди ақал TELEGRAM_BOT_TOKEN ва ADMIN_PASSWORD-ро
docker compose up -d --build
```

Сервер дар `http://localhost` (Nginx) ва `http://localhost:8080` (Go) кор мекунад.

### 3. Бе Docker (development)

```bash
cp .env.example .env
# базаи маълумотро роҳандозӣ кунед (PostgreSQL ва Redis)
make migrate-up
make run
```

### 4. Воридшавӣ

Аввалин корбар (admin) ҳангоми оғози сервер худкорона эҷод мешавад:

```
Login:    admin
Password: nav-xona@2026   (.env→ADMIN_PASSWORD)
```

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"nav-xona@2026"}'
```

---

## 📡 API ҳарчанди endpoint-ҳои асосӣ

> Ҳамаи endpoint-ҳо JWT-ро дар header `Authorization: Bearer <token>` интизор доранд (ба ғайр аз `/api/login` ва GET-ҳои оммавии бонкҳо/рассрочка).

| Method | URL                                          | Тавсиф                                  |
|--------|----------------------------------------------|-----------------------------------------|
| POST   | /api/login                                   | Воридшавӣ (返ад JWT)                    |
| POST   | /api/logout                                  | Хориҷшавӣ (token-ро blacklist мекунад)  |
| GET    | /api/check-auth                              | Тафтиши воридшавӣ                       |
| POST   | /api/me/telegram-link                        | Сохтани code барои пайвасткунии Telegram |
| GET    | /api/users                                   | Рӯйхати корбарон (admin)                |
| POST   | /api/users                                   | Корбари нав (admin) + tg-link code     |
| GET    | /api/tasks                                   | Вазифаҳо                                |
| POST   | /api/tasks                                   | Эҷоди вазифа (Telegram notify)         |
| GET    | /api/requests-boards                         | Доскаҳои заявка                         |
| POST   | /api/requests-new                            | Эҷоди заявка дар колонка                |
| POST   | /api/requests-move                           | Ҷой иваз кардани заявка                 |
| GET    | /api/kanban/boards                           | Лидҳои Kanban                           |
| POST   | /api/kanban/leads/move                       | Ҷой иваз кардани лид (Telegram notify) |
| GET    | /api/objekt/projects                         | Шахматка-и лоиҳаҳо                      |
| GET    | /api/objekt/export/excel?project_id=         | Excel-export-и шахматка                  |
| GET    | /api/sim-cards                               | SIM-кортҳо                              |
| POST   | /api/sim-tariffs/notify-expiring (admin)     | Огоҳнома ба Telegram дар бораи ба охир расидани тариф |
| GET    | /api/banks                                   | Бонкҳо (оммавӣ)                         |
| GET    | /api/installment-objects                     | Объектҳои рассрочка                     |
| GET    | /api/dashboard/stats                         | Омор                                    |

(*) Барои рӯйхати пурра ба `internal/router/router.go` нигаред.

---

## 🤖 Интегратсияи Telegram

### Чӣ тавр кор мекунад

1. Админ корбари нав эҷод мекунад (`POST /api/users`).
   Ҷавоб дорои майдонҳо:
   - `telegram_link_code` — рамзи якдафъаина (15 дақиқа)
   - `telegram_deep_link`  — ҳаволаи рост ба бот: `https://t.me/<bot>?start=<code>`

2. Корбари нав ҳаволаро мекушояд → бот ӯро мепайвандад →
   `users.telegram_chat_id` сабт мешавад.

3. Аз он лаҳза ба корбар тамоми огоҳсозиҳо мерасанд:
   - Ба ӯ вазифа таъин шуд (`tg.task_assigned`)
   - Заявка таъин шуд (`tg.request_assigned`)
   - Лиди ӯ ба колонкаи нав гузашт (`tg.lead_moved`)
   - Тарифи симкортааш ба охир мерасад (`tg.tariff_expiring`)

### Танзимот

`.env`:
```
TELEGRAM_BOT_TOKEN=<token аз @BotFather>
TELEGRAM_BOT_USERNAME=somon_crm_bot
TELEGRAM_NOTIFICATIONS_ENABLED=true
```

Барои хомӯш кардани огоҳсозиҳо ба як корбар:
```sql
UPDATE users SET notify_telegram = false WHERE id = ?;
```

### Фармонҳои бот

- `/start` — кӯмак
- `/start <CODE>` — пайвасткунӣ (deep-link)
- `/link <CODE>` — пайвасткунӣ (фармонӣ)
- `/help` — кӯмак

---

## 🌍 Чандзабонӣ (i18n)

Сервер дар header `Accept-Language` ё query `?lang=ru|tg` забонро муайян мекунад.
Забонҳои дастгиришаванда:

- `tg` — тоҷикӣ (default)
- `ru` — русӣ

Барои илова кардани забон:
1. Ба `internal/i18n/bundles.go` калидҳои нав илова кунед
2. Constant-и нав дар `i18n.go` (масалан `LocaleEN`)

---

## 🔒 Амният

- Паролҳо bcrypt-cost=12
- JWT HS256 + JTI + Redis blacklist дар logout
- Rate-limit барои /api/login — TODO (Flask-app маҷбур буд, ҷои Go ба `tollbooth` ё `gin-limiter`)
- File upload — extension whitelist, traversal-protection
- CORS — `CORS_ALLOWED_ORIGINS` дар `.env`
- Path-traversal — санҷиши `..` дар `/uploads/*`

---

## 🧪 Сохтан ва тест

```bash
make build       # Compile binaries (bin/server, bin/migrate)
make lint        # go vet + go fmt
make test        # go test ./... -race
make migrate-up  # Run SQL migrations
make migrate-down
```

---

## 📦 Production деплой

1. Танзимот:
   ```
   APP_ENV=production
   JWT_SECRET=<>=64 bytes random>
   ADMIN_PASSWORD=<unique>
   DB_PASSWORD=<unique>
   CORS_ALLOWED_ORIGINS=https://crm.example.com
   ```
2. `docker compose up -d --build`
3. SSL — пеши Nginx-и худатонро гузошта (Let's Encrypt) ё `certbot`-ро ба container-и nginx илова намоед.

---

## 🚀 Қадамҳои оянда (Frontend)

Frontend (Nuxt 4 + Vue 3 + TS + Tailwind + Pinia + i18n + SSR) дар як қадами навбатӣ
мутобиқан ба тарҳи зер сохта хоҳад шуд:

- `frontend/` — алоҳида
- API base URL: `${process.env.NUXT_PUBLIC_API_URL}` → backend
- JWT — дар httpOnly cookie + Pinia store
- SSR fetch бо `useFetch` барои дашборд ва саҳифаҳои хидмат

Frontend-ро дар як session-и навбатӣ хоҳем сохт пас аз муҳокима бо корбар.

---

## ✍️ Муаллиф

Лоиҳаи ислоҳшудаи `instaagrammeta/somon-crm` (бо ҳама функсияҳои нусхаи Flask
+ Telegram-bot integration).
