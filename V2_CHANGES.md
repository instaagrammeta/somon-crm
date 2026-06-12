# Somon CRM — v-2 changes

This branch (`v-2`) extends the v-1-соз baseline with the **Phase 3 (Realtime + 2026)** and **Phase 4 (Mobile + AI)** roadmap, plus a public-facing real-estate website.

## What is new

### 🌐 Public website (no auth)
- Marketing home, catalog, house detail, contact, about, install pages.
- Auto-popup that asks for name + phone after 30s on the site (once per session).
- Inline contact-form on every house detail card.
- Lead capture writes to a **separate** `website_leads` table; admins triage and "promote" qualified leads into the CRM pipeline.
- All public endpoints under `/api/public/*` and respect `public_visible` flags so CRM-only data never leaks.

### 🥽 Virtual tours (Marzipano)
- 360° equirectangular panoramas with Marzipano tile rendering (4K/8K-ready).
- Hotspots: `link` (jumps Kitchen → Living → Bedroom), `info` (text bubble), `url` (external).
- Admin CRUD for tours / panoramas / hotspots.
- Bottom strip room switcher + fullscreen toggle in the viewer.

### 📱 PWA + APK
- `@vite-pwa/nuxt` integration: installable Web App, offline cache for `/api/public/*` + uploads, Web Push manifest.
- `/install` page that sniffs platform: APK button on Android (served via `GET /api/public/apk`), "Add to Home Screen" instructions on iOS.

### ⚡ Realtime — WebSocket
- `gorilla/websocket` hub at `/api/ws?token=...`.
- Topics: `notif`, `chat`, `kanban`, `tasks`, `leads`, `presence`, `voice`, `ai`.
- Push: every `Notif.Push()` is mirrored to the user's open browser tabs (no F5).
- Online/offline indicator via the new `user_presence` table.

### 🤖 AI (DeepSeek by default)
DeepSeek's API is OpenAI-compatible, so the same client also works with OpenAI / Groq / Together.

| Endpoint | What it does |
|----------|--------------|
| `POST /api/ai/assistant` | free-form Q&A grounded in a server-built CRM snapshot |
| `POST /api/ai/score-lead/:id` | rate a website lead 0-100 + verdict + next-step |
| `POST /api/ai/summarize` | collapse text into 2-3 sentences |
| `POST /api/ai/categorize` | 3-6 lowercase tags |
| `POST /api/ai/ocr` | vision model: passport / contract → extracted fields |
| `POST /api/voice/upload` | audio → Whisper transcript (async, real-time push) |
| `GET /api/ai/logs` | admin: per-call token + duration audit |

Every AI call is recorded in `ai_logs` (tokens, duration, model, error).

### 🛡️ 2FA (TOTP, RFC 6238)
- Self-service enrollment at `/me/2fa`: QR + 6-digit code + 8 one-time recovery codes.
- `pquerna/otp` validates codes; recovery codes are stored as base32 hashes.

### 🧰 Workflow builder (no-code automation)
- Triggers: `lead.created`, `task.status`, `request.created`, `website_lead.created`, `cron.daily`.
- Step types: `notify_user`, `notify_role`, `tg_send`, `whatsapp_send`, `email_send`, `create_task`, `set_field`, `delay_minutes`, `log`.
- Each run is recorded in `workflow_runs` so the UI can show a history.

### 📜 Audit log
- Every mutating `/api/*` request stored in `audit_logs` with redacted payload, IP, UA, status.
- Admin-only read endpoint `GET /api/audit-logs?entity=&user_id=&limit=`.

### 💬 WhatsApp + Email
- WhatsApp Business Cloud API: GET handshake + POST inbound; outbound `SendText`/`SendTemplate`.
- SMTP outbound (TLS-aware); IMAP receive scaffolded for the next iteration.

## Database

New migration `0004_v2_realtime_ai`:
- `audit_logs`, `user_two_factors`, `website_leads`,
- `tours`, `panoramas`, `panorama_hotspots`,
- `voice_notes`, `ai_logs`,
- `workflows`, `workflow_runs`,
- `integration_channels`, `user_presence`.

Plus `houses` gets `public_visible`, `public_short_desc`, `public_full_desc`, `public_price_from`, `public_gallery`, `public_features`, `lat`, `lng`.

## Required env (see `.env.example`)
```
AI_API_KEY=sk-...                  # DeepSeek (or any OpenAI-compatible)
WHATSAPP_ENABLED=true              # +PHONE_NUMBER_ID + ACCESS_TOKEN + VERIFY_TOKEN
EMAIL_ENABLED=true                 # +SMTP creds
SITE_BRAND_NAME, SITE_PHONE, SITE_APK_URL, ...
```

## Routes summary

### Public
- `/` `/properties` `/properties/:id` `/tour/:slug` `/about` `/contact` `/install`
- `GET /api/public/site/config` `GET /api/public/houses` `GET /api/public/tours/:slug`
- `POST /api/public/leads` `GET /api/public/apk`

### CRM (auth)
- `/dashboard` (was `/`), the rest unchanged
- `/admin/website-leads` (admin), `/me/2fa`
- WS: `GET /api/ws`

### v-2 API
- `/api/2fa/{status,setup,verify,disable}`
- `/api/ai/{assistant,score-lead/:id,summarize,categorize,ocr,logs}`
- `/api/voice/{upload,list,...}`, `/api/tours`, `/api/panoramas`, `/api/hotspots`
- `/api/website-leads`, `/api/website-leads/:id/promote`
- `/api/workflows`, `/api/workflows/:id/runs`, `/api/audit-logs`
- `/api/webhooks/whatsapp` (Meta callbacks)

## Notes

- The sandbox has limited network access, so no `go build` / `npm install` was performed inside this PR. The Go side passes `gofmt -l` on every new file; the frontend uses standard Nuxt 4 conventions and the JSON locale files are validated.
