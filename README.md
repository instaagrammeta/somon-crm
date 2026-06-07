# Somon CRM

Реал тиҷорати ҳамтои CRM — пурра аз нав сохта шуда дар Go (backend) +
Nuxt 4 (frontend) + PostgreSQL + Redis + Telegram bot интегратсия.

> Бо як фармон ба кор медарояд: **`docker compose up -d --build`**

---

## ✨ Имкониятҳо

- **Дашборд** бо чартҳо ва оморот
- **Заявкаҳо** Kanban-доска бо drag&drop, columns, board export Excel
- **Лидҳо** Kanban (kanban_boards) бо ҳаракати лидҳо ва огоҳсозӣ
- **Вазифаҳо** бо иҷрокунанда ва ҳолатҳо
- **Шахматка** — матрицаи хонаҳо аз рӯи ошёна, статусҳо free/reserved/sold
- **Объектҳои манзил** ва бино бо координатаҳои харита
- **SIM-кортҳо** ва телефонҳои корпоративӣ + тарифҳо + огоҳсозии охирин рӯзҳо
- **Постҳо** бо метрикаҳо (likes, views, etc) + Excel export
- **Чат** — паёмҳои дохилӣ + файлҳои замимавӣ
- **Папкаҳои файлӣ** бо boole-ҳо
- **Ипотека** ва **Рассрочка** — бонкҳо/объектҳо + шартҳо (slug routing)
- **Корбарон** (admin only) бо Telegram-link автомат
- **🤖 Telegram bot пурра**:
  - ҳангоми эҷоди корбар автомат deep-link сохта мешавад
  - огоҳсозиҳо барои task/request/lead/sim
- **🌍 Чандзабонӣ** — Тоҷикӣ (default) + Русӣ

---

## 🏗 Tech stack

| Қабат          | Технология                                                    |
|----------------|---------------------------------------------------------------|
| **Frontend**   | Nuxt 4 + Vue 3 + TypeScript + Tailwind CSS + Pinia + i18n + SSR |
| **Backend**    | Go 1.23 + Gin + GORM + PostgreSQL + Redis                      |
| **Auth**       | JWT (HS256) + bcrypt + Redis blacklist                         |
| **Bot**        | Telegram Bot API (long-poll)                                   |
| **Infra**      | Docker + docker-compose + Nginx reverse proxy                  |
| **Migrations** | golang-migrate (SQL files)                                     |

---

## 🚀 Шурӯъ кардан

### 1. Клон + .env

```bash
git clone https://github.com/instaagrammeta/somon-crm.git
cd somon-crm
cp .env.example .env
# Дар .env инҳоро ҳатман иваз кунед:
#   - DB_PASSWORD
#   - JWT_SECRET (openssl rand -hex 32)
#   - ADMIN_PASSWORD
#   - TELEGRAM_BOT_TOKEN (агар бот лозим бошад)
nano .env
```

### 2. Docker compose up

```bash
docker compose up -d --build
```

Ҳамин ҳама — ҳама хидматҳо ба кор медароянд:

| Service       | Port | Тавсиф |
|---------------|------|--------|
| nginx         | 80   | Public reverse proxy |
| frontend      | -    | Nuxt SSR (internal) |
| backend       | -    | Go + Gin (internal) |
| postgres      | -    | DB (internal) |
| redis         | -    | Cache + JWT blacklist |

Ба `http://<server-ip>` (ё `http://localhost`) равед.

### 3. Воридшавии аввалин

```
Login:    admin
Password: <ADMIN_PASSWORD аз .env>
```

---

## 📁 Сохтор

```
somon-crm/
├── backend/                  # Go + Gin
│   ├── cmd/{server,migrate}/
│   ├── internal/
│   │   ├── config/           # viper + .env
│   │   ├── database/         # postgres + redis
│   │   ├── i18n/             # tg + ru bundles
│   │   ├── middleware/       # JWT, CORS, locale
│   │   ├── models/           # 30+ GORM models
│   │   ├── service/          # auth + telegram
│   │   ├── handler/          # 13 handler files, 180+ endpoints
│   │   ├── router/           # routing
│   │   └── utils/            # upload, excel, slug
│   ├── migrations/
│   ├── Dockerfile
│   └── README.md
├── frontend/                 # Nuxt 4 SSR
│   ├── components/           # UI + business components
│   ├── composables/          # useApi, useCrud, useToast
│   ├── stores/               # auth, ui (Pinia)
│   ├── pages/                # ~20 saҳифа
│   ├── layouts/              # default (sidebar) + auth
│   ├── i18n/locales/         # tg.json, ru.json
│   ├── middleware/           # auth.global.ts
│   ├── assets/css/           # Tailwind + base styles
│   ├── tailwind.config.ts
│   ├── nuxt.config.ts
│   └── Dockerfile
├── docker-compose.yml        # все сервисы
├── nginx.conf                # reverse proxy
├── .env.example
└── README.md (you are here)
```

---

## 🔧 Development (бе Docker)

### Backend

```bash
cd backend
cp .env.example .env
# Postgres + Redis-ро роҳандозӣ кунед (docker run...)
go mod tidy
go run ./cmd/server
```

### Frontend

```bash
cd frontend
cp .env.example .env
echo "NUXT_PUBLIC_API_BASE=http://localhost:8080" >> .env
npm install
npm run dev
```

`http://localhost:3000` ба кор медарояд.

---

## 🤖 Танзими Telegram-bot

1. Ба [@BotFather](https://t.me/BotFather) равед → `/newbot` → token гиред.
2. Дар `.env`:
   ```
   TELEGRAM_BOT_TOKEN=<token>
   TELEGRAM_BOT_USERNAME=<username бе @>
   TELEGRAM_NOTIFICATIONS_ENABLED=true
   ```
3. `docker compose restart backend`
4. Дар саҳифаи **Telegram** (sidebar → Telegram) рамзро гиред ва ба бот фиристед.
5. Ҳангоми эҷоди корбари нав, админ автомат deep-link мегирад — фақат бояд онро ба корбар фиристод.

---

## 🔐 Production checklist

- [ ] `JWT_SECRET` ҳадди ақал 32 байт random (`openssl rand -hex 32`)
- [ ] `DB_PASSWORD` ва `ADMIN_PASSWORD` сахт
- [ ] HTTPS (Let's Encrypt / certbot ё пушти Cloudflare)
- [ ] `APP_ENV=production`
- [ ] Backup-и regular барои PostgreSQL
- [ ] Tail of logs: `docker compose logs -f backend`

---

## 📜 Litsenziya

MIT (ё ҳар чизе ки шумо мехоҳед).

---

## ✍️ Authoring

Версияи Python (Flask) комилан ислоҳ карда шуд. Backend ба Go + Gin
кӯчонида, frontend ба Nuxt 4 SSR гузошта шуд. Бо Kiro дар як session
сохта шудааст.
