-- v-2: Realtime + AI + 2FA + Public website + Virtual tours + Workflows + Audit
-- Every table is "CREATE TABLE IF NOT EXISTS" so re-running this migration on a
-- partially-migrated database is safe.

-- ============================================================================
-- 1) Audit log: every mutating action by every user, used for compliance and
--    "who changed what when" timelines.
-- ============================================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id     INTEGER,
    user_login  VARCHAR(64),
    method      VARCHAR(10),
    path        VARCHAR(255),
    entity      VARCHAR(64),
    entity_id   VARCHAR(64),
    action      VARCHAR(32),
    ip          VARCHAR(64),
    user_agent  VARCHAR(255),
    status      INTEGER,
    payload     JSONB DEFAULT '{}'::jsonb
);
CREATE INDEX IF NOT EXISTS idx_audit_user      ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_created   ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_entity    ON audit_logs(entity, entity_id);

-- ============================================================================
-- 2) Two-factor authentication (TOTP). One row per user; presence with
--    enabled=true means the user must pass a 6-digit code on login.
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_two_factors (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret       VARCHAR(64) NOT NULL,
    enabled      BOOLEAN NOT NULL DEFAULT FALSE,
    recovery     JSONB DEFAULT '[]'::jsonb,  -- one-time recovery codes
    confirmed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 3) Public website leads. Captured by the auto-popup on the public site;
--    *separate* from the CRM `leads` table so the public form cannot pollute
--    the sales pipeline directly. An admin can promote a website_lead to a
--    real CRM lead from the dashboard.
-- ============================================================================
CREATE TABLE IF NOT EXISTS website_leads (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name            VARCHAR(128),
    phone           VARCHAR(32) NOT NULL,
    message         TEXT,
    source          VARCHAR(64),       -- "popup", "contact_form", "house_card", "tour"
    house_id        INTEGER REFERENCES houses(id) ON DELETE SET NULL,
    object_id       INTEGER REFERENCES objekts(id) ON DELETE SET NULL,
    page            VARCHAR(255),      -- url where the lead came from
    referrer        VARCHAR(255),
    user_agent      VARCHAR(255),
    ip              VARCHAR(64),
    utm_source      VARCHAR(64),
    utm_medium      VARCHAR(64),
    utm_campaign    VARCHAR(128),
    status          VARCHAR(32) NOT NULL DEFAULT 'new',  -- new|contacted|qualified|spam|promoted
    promoted_lead_id INTEGER,           -- id of the resulting CRM `leads` row
    notes           TEXT,
    ai_score        INTEGER DEFAULT 0,  -- 0..100 from AI lead scoring
    ai_summary      TEXT
);
CREATE INDEX IF NOT EXISTS idx_wleads_status   ON website_leads(status);
CREATE INDEX IF NOT EXISTS idx_wleads_created  ON website_leads(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wleads_house    ON website_leads(house_id);

-- ============================================================================
-- 4) Public-facing extra fields on houses (used by the website).
-- ============================================================================
ALTER TABLE houses ADD COLUMN IF NOT EXISTS public_visible    BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE houses ADD COLUMN IF NOT EXISTS public_short_desc VARCHAR(500);
ALTER TABLE houses ADD COLUMN IF NOT EXISTS public_full_desc  TEXT;
ALTER TABLE houses ADD COLUMN IF NOT EXISTS public_price_from BIGINT;
ALTER TABLE houses ADD COLUMN IF NOT EXISTS public_gallery    JSONB DEFAULT '[]'::jsonb;
ALTER TABLE houses ADD COLUMN IF NOT EXISTS public_features   JSONB DEFAULT '[]'::jsonb;
ALTER TABLE houses ADD COLUMN IF NOT EXISTS lat               DOUBLE PRECISION;
ALTER TABLE houses ADD COLUMN IF NOT EXISTS lng               DOUBLE PRECISION;

-- ============================================================================
-- 5) Virtual tours: a tour belongs to a house (or stand-alone), and contains a
--    set of panoramas (HDRI / equirectangular images). Each panorama has a set
--    of hotspots that link to the next panorama (Kitchen → Living → Bedroom).
-- ============================================================================
CREATE TABLE IF NOT EXISTS tours (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    title           VARCHAR(255) NOT NULL,
    slug            VARCHAR(128) UNIQUE,
    description     TEXT,
    house_id        INTEGER REFERENCES houses(id) ON DELETE SET NULL,
    object_id       INTEGER REFERENCES objekts(id) ON DELETE SET NULL,
    cover_image     VARCHAR(512),
    public_visible  BOOLEAN NOT NULL DEFAULT TRUE,
    initial_panorama_id INTEGER     -- start scene
);

CREATE TABLE IF NOT EXISTS panoramas (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tour_id         INTEGER NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    title           VARCHAR(128) NOT NULL,           -- "Ошхона", "Меҳмонхона"
    image_url       VARCHAR(512) NOT NULL,           -- equirectangular .jpg / hdr
    image_type      VARCHAR(32) NOT NULL DEFAULT 'equirect',  -- equirect|cube|hdr
    initial_yaw     DOUBLE PRECISION DEFAULT 0,      -- starting view (radians)
    initial_pitch   DOUBLE PRECISION DEFAULT 0,
    initial_zoom    DOUBLE PRECISION DEFAULT 90,     -- field of view (deg)
    sort_order      INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_panorama_tour ON panoramas(tour_id);

-- A hotspot lives on a panorama and either jumps to another panorama
-- (kind='link') or shows an info bubble (kind='info').
CREATE TABLE IF NOT EXISTS panorama_hotspots (
    id              SERIAL PRIMARY KEY,
    panorama_id     INTEGER NOT NULL REFERENCES panoramas(id) ON DELETE CASCADE,
    target_panorama_id INTEGER REFERENCES panoramas(id) ON DELETE SET NULL,
    kind            VARCHAR(16) NOT NULL DEFAULT 'link',  -- link|info|url
    label           VARCHAR(128),
    yaw             DOUBLE PRECISION NOT NULL,            -- radians
    pitch           DOUBLE PRECISION NOT NULL,
    info_text       TEXT,
    info_url        VARCHAR(512)
);
CREATE INDEX IF NOT EXISTS idx_hotspot_panorama ON panorama_hotspots(panorama_id);

-- Add the FK on tours.initial_panorama_id only after panoramas exists.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
         WHERE constraint_name = 'tours_initial_panorama_fk'
    ) THEN
        ALTER TABLE tours
            ADD CONSTRAINT tours_initial_panorama_fk
            FOREIGN KEY (initial_panorama_id) REFERENCES panoramas(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================================
-- 6) Voice notes attached to tasks / leads / requests. The audio file lives on
--    disk (uploads/voice/), the row stores its path + the AI transcript.
-- ============================================================================
CREATE TABLE IF NOT EXISTS voice_notes (
    id              SERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id         INTEGER REFERENCES users(id) ON DELETE SET NULL,
    entity          VARCHAR(32) NOT NULL,    -- "task" | "lead" | "request" | "chat"
    entity_id       INTEGER NOT NULL,
    audio_url       VARCHAR(512) NOT NULL,
    duration_sec    INTEGER DEFAULT 0,
    transcript      TEXT,
    transcript_lang VARCHAR(8),
    status          VARCHAR(16) NOT NULL DEFAULT 'pending'  -- pending|done|failed
);
CREATE INDEX IF NOT EXISTS idx_voice_entity ON voice_notes(entity, entity_id);

-- ============================================================================
-- 7) AI request log — every call to the AI provider, used for billing
--    accountability, rate-limiting, and showing "AI history" to the user.
-- ============================================================================
CREATE TABLE IF NOT EXISTS ai_logs (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id         INTEGER REFERENCES users(id) ON DELETE SET NULL,
    feature         VARCHAR(32) NOT NULL,    -- assistant|score|summarize|ocr|whisper|categorize
    model           VARCHAR(64),
    prompt_tokens   INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens    INTEGER DEFAULT 0,
    duration_ms     INTEGER DEFAULT 0,
    request_brief   TEXT,
    response_brief  TEXT,
    error           TEXT
);
CREATE INDEX IF NOT EXISTS idx_ai_logs_user_feature ON ai_logs(user_id, feature);

-- ============================================================================
-- 8) Workflow builder. A workflow has a trigger (e.g. "lead created in column X")
--    and an ordered list of steps (notify, change_field, create_task, send_tg).
--    Steps are stored as JSON for flexibility and edited from the UI builder.
-- ============================================================================
CREATE TABLE IF NOT EXISTS workflows (
    id          SERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name        VARCHAR(128) NOT NULL,
    description TEXT,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    trigger     VARCHAR(64) NOT NULL,    -- lead.created|lead.column_changed|task.status|request.created|cron.daily
    config      JSONB NOT NULL DEFAULT '{}'::jsonb,  -- trigger-specific filter (e.g. column id)
    steps       JSONB NOT NULL DEFAULT '[]'::jsonb,  -- list of {type, params}
    last_run_at TIMESTAMPTZ,
    last_status VARCHAR(16),
    runs        INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_workflows_trigger ON workflows(trigger) WHERE enabled = TRUE;

CREATE TABLE IF NOT EXISTS workflow_runs (
    id          BIGSERIAL PRIMARY KEY,
    workflow_id INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMPTZ,
    status      VARCHAR(16) NOT NULL DEFAULT 'running',
    payload     JSONB DEFAULT '{}'::jsonb,
    log         TEXT
);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_wf ON workflow_runs(workflow_id);

-- ============================================================================
-- 9) Misc: external messaging integration accounts (WhatsApp / Email / TG bot
--    can already exist). One row per channel; the actual message stream lives
--    in the existing `messages` table.
-- ============================================================================
CREATE TABLE IF NOT EXISTS integration_channels (
    id          SERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type        VARCHAR(32) NOT NULL,    -- whatsapp|email|telegram_bot
    name        VARCHAR(128),
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    config      JSONB NOT NULL DEFAULT '{}'::jsonb  -- credentials & settings (encrypted at rest is recommended)
);

-- ============================================================================
-- 10) Realtime presence (optional, used by the WS hub for "who is online").
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_presence (
    user_id     INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    online      BOOLEAN NOT NULL DEFAULT FALSE
);
