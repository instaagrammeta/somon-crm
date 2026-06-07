-- Somon CRM initial schema (PostgreSQL)
-- All timestamps are TIMESTAMPTZ.
-- All "files"/"phones" multi-value columns use jsonb.

CREATE TABLE IF NOT EXISTS users (
    id                  BIGSERIAL PRIMARY KEY,
    full_name           TEXT,
    age                 INT,
    personal_phones     JSONB DEFAULT '[]'::jsonb,
    work_phones         JSONB DEFAULT '[]'::jsonb,
    login               TEXT UNIQUE NOT NULL,
    password            TEXT NOT NULL,
    photo               TEXT DEFAULT '',
    category            TEXT DEFAULT '',
    role                TEXT DEFAULT 'employee',
    telegram_username   TEXT DEFAULT '',
    telegram_chat_id    BIGINT DEFAULT 0,
    telegram_linked_at  TIMESTAMPTZ,
    notify_telegram     BOOLEAN DEFAULT TRUE,
    is_active           BOOLEAN DEFAULT TRUE,
    last_login          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_users_role ON users(role);
CREATE INDEX IF NOT EXISTS ix_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS ix_users_telegram ON users(telegram_chat_id);

CREATE TABLE IF NOT EXISTS tasks (
    id            BIGSERIAL PRIMARY KEY,
    title         TEXT NOT NULL,
    description   TEXT,
    author_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    executor_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    photo         TEXT,
    status        TEXT DEFAULT 'new',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_tasks_executor ON tasks(executor_id);
CREATE INDEX IF NOT EXISTS ix_tasks_status ON tasks(status);

CREATE TABLE IF NOT EXISTS legacy_requests (
    id              BIGSERIAL PRIMARY KEY,
    property_type   TEXT,
    address         TEXT,
    area            DOUBLE PRECISION,
    rooms           INT,
    windows         INT,
    floor           INT,
    total_floors    INT,
    documents       TEXT,
    total_price     DOUBLE PRECISION,
    price_per_m2    DOUBLE PRECISION,
    phone           TEXT,
    client_name     TEXT,
    manager         TEXT,
    smm             TEXT,
    comment         TEXT,
    author_id       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    executor_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    files           JSONB DEFAULT '[]'::jsonb,
    status          TEXT DEFAULT 'new',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS posts (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    user_name       TEXT,
    title           TEXT,
    description     TEXT,
    category        TEXT,
    content_type    TEXT,
    project         TEXT,
    media_path      TEXT,
    media_type      TEXT,
    link            TEXT,
    post_date       TIMESTAMPTZ,
    likes           INT DEFAULT 0,
    comments        INT DEFAULT 0,
    shares          INT DEFAULT 0,
    views           INT DEFAULT 0,
    reach           INT DEFAULT 0,
    is_published    BOOLEAN DEFAULT TRUE,
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_posts_category ON posts(category);

CREATE TABLE IF NOT EXISTS messages (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    user_name   TEXT,
    message     TEXT,
    file_path   TEXT,
    file_name   TEXT,
    file_type   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS requests_boards (
    id          BIGSERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    color       TEXT DEFAULT '#0f172a',
    is_public   BOOLEAN DEFAULT TRUE,
    author_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS requests_columns (
    id          BIGSERIAL PRIMARY KEY,
    board_id    BIGINT REFERENCES requests_boards(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    color       TEXT DEFAULT '#3b82f6',
    order_index INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_req_cols_board ON requests_columns(board_id);

CREATE TABLE IF NOT EXISTS requests_items (
    id              BIGSERIAL PRIMARY KEY,
    column_id       BIGINT REFERENCES requests_columns(id) ON DELETE CASCADE,
    board_id        BIGINT REFERENCES requests_boards(id) ON DELETE CASCADE,
    property_type   TEXT,
    address         TEXT,
    area            DOUBLE PRECISION,
    rooms           INT,
    windows         INT,
    floor           INT,
    total_floors    INT,
    total_price     DOUBLE PRECISION,
    price_per_m2    DOUBLE PRECISION,
    phone           TEXT,
    client_name     TEXT,
    comment         TEXT,
    author_id       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    author_name     TEXT,
    executor_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    executor_name   TEXT,
    files           JSONB DEFAULT '[]'::jsonb,
    order_index     INT DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_req_items_board ON requests_items(board_id);
CREATE INDEX IF NOT EXISTS ix_req_items_column ON requests_items(column_id);
CREATE INDEX IF NOT EXISTS ix_req_items_order ON requests_items(column_id, order_index);

CREATE TABLE IF NOT EXISTS company_phones (
    id           BIGSERIAL PRIMARY KEY,
    model        TEXT NOT NULL,
    phone_id     TEXT UNIQUE,
    assigned_to  BIGINT REFERENCES users(id) ON DELETE SET NULL,
    description  TEXT,
    status       TEXT DEFAULT 'free',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sim_cards (
    id            BIGSERIAL PRIMARY KEY,
    phone_number  TEXT UNIQUE NOT NULL,
    operator      TEXT NOT NULL,
    assigned_to   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    phone_id      BIGINT REFERENCES company_phones(id) ON DELETE SET NULL,
    description   TEXT,
    status        TEXT DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_sim_assigned ON sim_cards(assigned_to);

CREATE TABLE IF NOT EXISTS sim_tariffs (
    id          BIGSERIAL PRIMARY KEY,
    sim_id      BIGINT REFERENCES sim_cards(id) ON DELETE CASCADE,
    minutes     INT DEFAULT 0,
    gb          DOUBLE PRECISION DEFAULT 0,
    sms         INT DEFAULT 0,
    cost        DOUBLE PRECISION DEFAULT 0,
    start_date  TIMESTAMPTZ,
    end_date    TIMESTAMPTZ,
    status      TEXT DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_sim_tariffs_sim ON sim_tariffs(sim_id);
CREATE INDEX IF NOT EXISTS ix_sim_tariffs_end ON sim_tariffs(end_date);

CREATE TABLE IF NOT EXISTS tariff_payments (
    id            BIGSERIAL PRIMARY KEY,
    sim_id        BIGINT REFERENCES sim_cards(id) ON DELETE CASCADE,
    tariff_id     BIGINT REFERENCES sim_tariffs(id) ON DELETE SET NULL,
    amount        DOUBLE PRECISION DEFAULT 0,
    payment_date  TIMESTAMPTZ,
    start_date    TIMESTAMPTZ,
    end_date      TIMESTAMPTZ,
    status        TEXT DEFAULT 'paid',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS houses (
    id                     BIGSERIAL PRIMARY KEY,
    title                  TEXT,
    construction_type      TEXT,
    district               TEXT,
    address                TEXT,
    area                   DOUBLE PRECISION,
    rooms                  INT,
    windows                INT,
    floor                  INT,
    total_floors           INT,
    price_per_m2           DOUBLE PRECISION,
    total_price            DOUBLE PRECISION,
    developer              TEXT,
    contact_phone          TEXT,
    has_tech_passport      TEXT DEFAULT 'нет',
    has_renovation_permit  TEXT DEFAULT 'нет',
    files                  JSONB DEFAULT '[]'::jsonb,
    author_id              BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS realty_objects (
    id                     BIGSERIAL PRIMARY KEY,
    name                   TEXT NOT NULL,
    description            TEXT,
    address                TEXT,
    lat                    DOUBLE PRECISION DEFAULT 38.5598,
    lng                    DOUBLE PRECISION DEFAULT 68.7870,
    construction_type      TEXT DEFAULT 'новостройка',
    district               TEXT DEFAULT 'н.Сино',
    area                   DOUBLE PRECISION DEFAULT 0,
    rooms                  INT DEFAULT 0,
    windows                INT DEFAULT 0,
    floor                  INT DEFAULT 0,
    total_floors           INT DEFAULT 0,
    price_per_m2           DOUBLE PRECISION DEFAULT 0,
    total_price            DOUBLE PRECISION DEFAULT 0,
    developer              TEXT,
    contact_phone          TEXT,
    has_tech_passport      TEXT DEFAULT 'нет',
    has_renovation_permit  TEXT DEFAULT 'нет',
    author_id              BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS realty_blocks (
    id          BIGSERIAL PRIMARY KEY,
    object_id   BIGINT REFERENCES realty_objects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    code        TEXT,
    order_index INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS realty_pricing (
    id                BIGSERIAL PRIMARY KEY,
    object_id         BIGINT REFERENCES realty_objects(id) ON DELETE CASCADE,
    block_id          BIGINT REFERENCES realty_blocks(id) ON DELETE CASCADE,
    floor_number      INT,
    floor_range_start INT,
    floor_range_end   INT,
    percent_value     INT,
    price_usd         DOUBLE PRECISION DEFAULT 0,
    price_tjs         DOUBLE PRECISION DEFAULT 0,
    currency          TEXT DEFAULT 'USD',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS realty_layouts (
    id            BIGSERIAL PRIMARY KEY,
    object_id     BIGINT REFERENCES realty_objects(id) ON DELETE CASCADE,
    room_type     TEXT,
    windows_count INT,
    area          DOUBLE PRECISION,
    price_usd     DOUBLE PRECISION DEFAULT 0,
    price_tjs     DOUBLE PRECISION DEFAULT 0,
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS lids (
    id           BIGSERIAL PRIMARY KEY,
    client_name  TEXT,
    phone        TEXT,
    topic        TEXT,
    comment      TEXT,
    source       TEXT,
    mortgage     BOOLEAN DEFAULT FALSE,
    box          BOOLEAN DEFAULT FALSE,
    author_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS folders (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT,
    parent_id   BIGINT DEFAULT 0,
    author_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    author_name TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_folders_parent ON folders(parent_id);

CREATE TABLE IF NOT EXISTS kanban_boards (
    id          BIGSERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    color       TEXT DEFAULT '#0079bf',
    is_public   BOOLEAN DEFAULT TRUE,
    author_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kanban_board_members (
    id         BIGSERIAL PRIMARY KEY,
    board_id   BIGINT REFERENCES kanban_boards(id) ON DELETE CASCADE,
    user_id    BIGINT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kanban_columns (
    id          BIGSERIAL PRIMARY KEY,
    board_id    BIGINT REFERENCES kanban_boards(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    color       TEXT DEFAULT '#0079bf',
    order_index INT DEFAULT 0,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kanban_leads (
    id          BIGSERIAL PRIMARY KEY,
    column_id   BIGINT REFERENCES kanban_columns(id) ON DELETE CASCADE,
    board_id    BIGINT REFERENCES kanban_boards(id) ON DELETE CASCADE,
    client_name TEXT,
    phone       TEXT,
    topic       TEXT,
    comment     TEXT,
    source      TEXT,
    mortgage    BOOLEAN DEFAULT FALSE,
    box         BOOLEAN DEFAULT FALSE,
    author_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    author_name TEXT,
    order_index INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_kleads_col ON kanban_leads(column_id, order_index);
CREATE INDEX IF NOT EXISTS ix_kleads_board ON kanban_leads(board_id);

CREATE TABLE IF NOT EXISTS kanban_lead_interactions (
    id           BIGSERIAL PRIMARY KEY,
    lead_id      BIGINT REFERENCES kanban_leads(id) ON DELETE CASCADE,
    phone        TEXT,
    topic        TEXT,
    source       TEXT,
    contact_type TEXT,
    comment      TEXT,
    mortgage     BOOLEAN DEFAULT FALSE,
    box          BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS folder_files (
    id            BIGSERIAL PRIMARY KEY,
    folder_id     BIGINT REFERENCES folders(id) ON DELETE CASCADE,
    filename      TEXT,
    original_name TEXT,
    filepath      TEXT,
    filetype      TEXT,
    filesize      BIGINT,
    author_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    author_name   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS objekt_projects (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    address     TEXT,
    developer   TEXT,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS objekt_blocks (
    id                     BIGSERIAL PRIMARY KEY,
    project_id             BIGINT REFERENCES objekt_projects(id) ON DELETE CASCADE,
    name                   TEXT NOT NULL,
    floor_from             INT NOT NULL,
    floor_to               INT NOT NULL,
    default_area           DOUBLE PRECISION NOT NULL,
    default_rooms          INT NOT NULL,
    default_windows        INT NOT NULL,
    default_price_per_m2   DOUBLE PRECISION NOT NULL,
    default_balcony        BOOLEAN DEFAULT FALSE,
    default_bathroom_type  TEXT DEFAULT 'combined',
    default_bathroom_count INT DEFAULT 1,
    default_plan_image     TEXT,
    description            TEXT,
    order_index            INT DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS objekt_apartments (
    id              BIGSERIAL PRIMARY KEY,
    block_id        BIGINT REFERENCES objekt_blocks(id) ON DELETE CASCADE,
    floor           INT NOT NULL,
    area            DOUBLE PRECISION NOT NULL,
    rooms           INT NOT NULL,
    windows         INT NOT NULL,
    price_per_m2    DOUBLE PRECISION NOT NULL,
    total_price     DOUBLE PRECISION NOT NULL,
    status          TEXT DEFAULT 'free',
    balcony         BOOLEAN DEFAULT FALSE,
    bathroom_type   TEXT DEFAULT 'combined',
    bathroom_count  INT DEFAULT 1,
    plan_image      TEXT,
    description     TEXT,
    client_name     TEXT,
    client_phone    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(block_id, floor)
);
CREATE INDEX IF NOT EXISTS ix_objekt_apt_status ON objekt_apartments(status);

CREATE TABLE IF NOT EXISTS banks (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    logo        TEXT,
    description TEXT,
    order_index INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mortgage_conditions (
    id                    BIGSERIAL PRIMARY KEY,
    bank_id               BIGINT REFERENCES banks(id) ON DELETE CASCADE,
    currency              TEXT NOT NULL DEFAULT 'TJS',
    interest_yearly       DOUBLE PRECISION DEFAULT 0,
    interest_monthly      DOUBLE PRECISION DEFAULT 0,
    min_amount            DOUBLE PRECISION DEFAULT 0,
    max_amount            DOUBLE PRECISION DEFAULT 0,
    min_months            INT DEFAULT 12,
    max_months            INT DEFAULT 240,
    down_payment_percent  DOUBLE PRECISION DEFAULT 30,
    guarantor_required    BOOLEAN DEFAULT FALSE,
    collateral_required   BOOLEAN DEFAULT FALSE,
    extra_conditions      TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS installment_objects (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    image       TEXT,
    description TEXT,
    address     TEXT,
    developer   TEXT,
    order_index INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS installment_conditions (
    id                    BIGSERIAL PRIMARY KEY,
    object_id             BIGINT REFERENCES installment_objects(id) ON DELETE CASCADE,
    title                 TEXT,
    currency              TEXT DEFAULT 'TJS',
    min_price             DOUBLE PRECISION DEFAULT 0,
    max_price             DOUBLE PRECISION DEFAULT 0,
    down_payment_percent  DOUBLE PRECISION DEFAULT 30,
    months                INT DEFAULT 12,
    monthly_payment_rule  TEXT,
    interest_percent      DOUBLE PRECISION DEFAULT 0,
    extra_conditions      TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
