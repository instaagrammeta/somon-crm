#!/usr/bin/env bash
# ============================================================================
# Somon CRM — one-command bootstrap.
#
# Usage:
#   ./bootstrap.sh                    # first install or update
#   ./bootstrap.sh --rebuild          # force a clean rebuild
#   ./bootstrap.sh --reset-admin      # reset admin password from .env once
#   ./bootstrap.sh --logs             # show backend logs once running
#
# What it does:
#   1. Verifies Docker + Compose are present.
#   2. Creates .env from .env.example if missing, generating fresh secrets.
#   3. Ensures ./static/apk exists (APK landing).
#   4. Builds and starts the stack (docker compose up -d --build).
#   5. Waits for the backend health endpoint to flip to "ok":true.
#   6. Prints the URL and admin login.
# ============================================================================
set -euo pipefail

cd "$(dirname "$0")"

color_green="\033[0;32m"; color_red="\033[0;31m"; color_yellow="\033[1;33m"; color_off="\033[0m"
say()  { printf "${color_green}==>${color_off} %s\n" "$*"; }
warn() { printf "${color_yellow}!! %s${color_off}\n" "$*"; }
fail() { printf "${color_red}xx %s${color_off}\n" "$*"; exit 1; }

# --- 1. Docker check -------------------------------------------------------
command -v docker >/dev/null 2>&1 || fail "docker is not installed."
if docker compose version >/dev/null 2>&1; then
    DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
    DC="docker-compose"
else
    fail "Neither 'docker compose' nor 'docker-compose' is available."
fi

REBUILD=0
LOGS=0
RESET_ADMIN=0
for a in "$@"; do
    case "$a" in
        --rebuild)      REBUILD=1 ;;
        --logs)         LOGS=1 ;;
        --reset-admin)  RESET_ADMIN=1 ;;
        -h|--help)
            sed -n '2,18p' "$0"; exit 0 ;;
    esac
done

# --- 2. .env ---------------------------------------------------------------
if [[ ! -f .env ]]; then
    say "Creating .env from .env.example with fresh secrets…"
    cp .env.example .env

    # Replace JWT_SECRET, DB_PASSWORD, ADMIN_PASSWORD with random values.
    rand() { openssl rand -hex 16 2>/dev/null || head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 32; }
    JWT=$(openssl rand -hex 32 2>/dev/null || rand)
    DB_PW=$(rand)
    ADMIN_PW=$(rand)
    # GNU sed and BSD sed both accept this form when -i has an empty backup arg.
    if sed --version >/dev/null 2>&1; then
        SED_INPLACE=(sed -i)
    else
        SED_INPLACE=(sed -i '')
    fi
    "${SED_INPLACE[@]}" "s|^JWT_SECRET=.*|JWT_SECRET=$JWT|"           .env
    "${SED_INPLACE[@]}" "s|^DB_PASSWORD=.*|DB_PASSWORD=$DB_PW|"       .env
    "${SED_INPLACE[@]}" "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD=$ADMIN_PW|" .env
    say "Generated secrets and saved to .env. ${color_yellow}Save the admin password:${color_off} $ADMIN_PW"
else
    say ".env already exists — leaving it alone."
fi

# --- 3. Static dir ---------------------------------------------------------
mkdir -p static/apk
[[ -f static/apk/.gitkeep ]] || touch static/apk/.gitkeep
say "static/apk/ ready (drop your APK file here)."

# --- 4. (re)build & launch -------------------------------------------------
if [[ $RESET_ADMIN -eq 1 ]]; then
    warn "Forcing admin password reset from .env on next start."
    if sed --version >/dev/null 2>&1; then
        sed -i  's|^RESET_ADMIN_PASSWORD=.*|RESET_ADMIN_PASSWORD=true|' .env
    else
        sed -i '' 's|^RESET_ADMIN_PASSWORD=.*|RESET_ADMIN_PASSWORD=true|' .env
    fi
fi

say "Starting the stack…"
if [[ $REBUILD -eq 1 ]]; then
    $DC down --remove-orphans
    $DC build --no-cache
fi
$DC up -d --build

# --- 5. Wait for backend health -------------------------------------------
say "Waiting for backend to become healthy (max 180s)…"
for i in $(seq 1 90); do
    state=$($DC ps backend --format json 2>/dev/null | grep -o '"Health":"[^"]*"' | head -1 | cut -d'"' -f4 || true)
    if [[ "$state" == "healthy" ]]; then
        say "Backend is healthy."
        break
    fi
    sleep 2
done

# --- 6. Final notes --------------------------------------------------------
PORT=$(grep -E '^PUBLIC_PORT=' .env | head -1 | cut -d= -f2 || echo 80)
PORT=${PORT:-80}
ADMIN_LOGIN=$(grep -E '^ADMIN_LOGIN=' .env | head -1 | cut -d= -f2)
HOST=$(hostname -I 2>/dev/null | awk '{print $1}' || echo localhost)

cat <<EOF

${color_green}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${color_off}
  ${color_green}Somon CRM аст тайёр!${color_off}

  Public website : http://$HOST:$PORT/
  CRM login      : http://$HOST:$PORT/login        ($ADMIN_LOGIN / from .env)
  Health         : http://$HOST:$PORT/health
  Logs           : $DC logs -f backend
  Stop           : $DC down
${color_green}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${color_off}

EOF

if [[ $RESET_ADMIN -eq 1 ]]; then
    warn "Don't forget to set RESET_ADMIN_PASSWORD=false in .env after this start."
fi

if [[ $LOGS -eq 1 ]]; then
    $DC logs -f backend
fi
