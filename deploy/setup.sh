#!/bin/bash
# First-time setup on a fresh Ubuntu/Debian VPS (run as root).
# Safe to re-run — all steps are idempotent.
# Usage: bash deploy/setup.sh

set -e
REPO_URL="https://github.com/888Greys/smm.git"
REPO_DIR="/opt/smm/repo"
BIN_DIR="/opt/smm/bin"
ENV_FILE="/opt/smm/.env"
DB_USER="smm"
DB_NAME="smm"
DB_PASS="changeme"   # change this in .env after setup

echo "=== SMM Panel First-Time Setup ==="

# ── 1. System packages ────────────────────────────────────────────────────────
echo "[1/9] Installing system packages..."
apt update -qq
apt install -y git curl wget nginx certbot python3-certbot-nginx postgresql

# ── 2. Go (install 1.22 from official source if not present / too old) ────────
echo "[2/9] Checking Go installation..."
GO_MIN="1.21"
install_go() {
  echo "      Installing Go 1.22..."
  GO_TAR="go1.22.4.linux-amd64.tar.gz"
  wget -q "https://go.dev/dl/${GO_TAR}" -O /tmp/go.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
  chmod +x /etc/profile.d/go.sh
}
export PATH=$PATH:/usr/local/go/bin
if ! command -v go &>/dev/null; then
  install_go
elif ! go version | grep -qE "go1\.(2[1-9]|[3-9][0-9])"; then
  install_go
fi
export PATH=$PATH:/usr/local/go/bin
echo "      $(go version)"

# ── 3. System user ────────────────────────────────────────────────────────────
echo "[3/9] Creating smm system user..."
useradd -r -m -d /opt/smm -s /bin/bash smm 2>/dev/null || true

# ── 4. Clone / pull repo ──────────────────────────────────────────────────────
echo "[4/9] Cloning repository..."
if [ -d "$REPO_DIR/.git" ]; then
  git -C "$REPO_DIR" pull
else
  git clone "$REPO_URL" "$REPO_DIR"
fi

# ── 5. Build binaries ─────────────────────────────────────────────────────────
echo "[5/9] Building binaries..."
mkdir -p "$BIN_DIR"
cd "$REPO_DIR"
go mod tidy
go build -o "$BIN_DIR/bot"    ./cmd/bot
go build -o "$BIN_DIR/worker" ./cmd/worker
go build -o "$BIN_DIR/server" ./cmd/server
chown -R smm:smm "$BIN_DIR"
echo "      Binaries built: bot, worker, server"

# ── 6. PostgreSQL ─────────────────────────────────────────────────────────────
echo "[6/9] Configuring PostgreSQL..."
# Create user
sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASS}';"
# Create database
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE DATABASE ${DB_NAME} OWNER ${DB_USER};"
# Apply schema (ignore errors on re-run — tables already exist)
sudo -u postgres psql -d "$DB_NAME" -f "$REPO_DIR/internal/db/schema.sql" 2>/dev/null || true
# Apply v2 migration (IF NOT EXISTS — safe to re-run)
sudo -u postgres psql -d "$DB_NAME" -f "$REPO_DIR/internal/db/migrate_v2.sql" 2>&1 | grep -v "^$" || true
# Grant permissions on all current and future objects
sudo -u postgres psql -d "$DB_NAME" -c "
  GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};
  GRANT ALL PRIVILEGES ON ALL TABLES    IN SCHEMA public TO ${DB_USER};
  GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${DB_USER};
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES    TO ${DB_USER};
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ${DB_USER};
" > /dev/null
echo "      Database ready"

# ── 7. .env file ──────────────────────────────────────────────────────────────
echo "[7/9] Configuring environment..."
if [ ! -f "$ENV_FILE" ]; then
  cp "$REPO_DIR/.env.example" "$ENV_FILE"
  # Set the real DB password
  sed -i "s|postgres://smm:changeme@|postgres://${DB_USER}:${DB_PASS}@|" "$ENV_FILE"
  chown smm:smm "$ENV_FILE"
  chmod 600 "$ENV_FILE"
  echo ""
  echo "  ┌─────────────────────────────────────────────────────┐"
  echo "  │  IMPORTANT: Fill in your API keys in /opt/smm/.env  │"
  echo "  │  before starting the services.                       │"
  echo "  │                                                       │"
  echo "  │  nano /opt/smm/.env                                  │"
  echo "  └─────────────────────────────────────────────────────┘"
  echo ""
else
  echo "      .env already exists — skipping"
fi

# ── 8. Nginx ──────────────────────────────────────────────────────────────────
echo "[8/9] Configuring nginx..."

# api.innbucks.org → Go API server (port 8005)
cp "$REPO_DIR/deploy/nginx/api.innbucks.org.conf" /etc/nginx/sites-available/api.innbucks.org
ln -sf /etc/nginx/sites-available/api.innbucks.org /etc/nginx/sites-enabled/api.innbucks.org
rm -f /etc/nginx/sites-enabled/default

nginx -t && systemctl reload nginx

# SSL for api subdomain
certbot --nginx -d api.innbucks.org --non-interactive --agree-tos -m admin@innbucks.org 2>/dev/null || \
  echo "      Certbot: run manually → certbot --nginx -d api.innbucks.org"

echo "      Nginx ready"

# ── 9. Systemd services ───────────────────────────────────────────────────────
echo "[9/9] Installing systemd services..."
cp "$REPO_DIR/deploy/systemd/smm-bot.service"    /etc/systemd/system/
cp "$REPO_DIR/deploy/systemd/smm-worker.service" /etc/systemd/system/
cp "$REPO_DIR/deploy/systemd/smm-server.service" /etc/systemd/system/
systemctl daemon-reload
systemctl enable smm-bot smm-worker smm-server

echo ""
echo "=== Setup complete ==="
echo ""
echo "Next steps:"
echo "  1. Edit your keys:  nano /opt/smm/.env"
echo "  2. Start services:  systemctl start smm-bot smm-worker smm-server"
echo "  3. Check status:    systemctl status smm-bot smm-worker smm-server"
echo "  4. Watch logs:      journalctl -u smm-bot -u smm-worker -u smm-server -f"
echo ""
echo "For future updates:  bash /opt/smm/repo/deploy/update.sh"
echo ""
echo "Frontend (innbucks.org) → deploy on Vercel, root dir: web/"
echo "  NEXT_PUBLIC_API_URL=https://api.innbucks.org"
echo ""
