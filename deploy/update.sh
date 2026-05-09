#!/bin/bash
# Quick update: pull latest code, rebuild, run migrations, restart.
# Usage (on server as root): bash /opt/smm/repo/deploy/update.sh

set -e
cd /opt/smm/repo

echo "=== SMM Panel Update ==="

# 1. Pull
echo "[1/4] Pulling latest code..."
git pull

# 2. Build
echo "[2/4] Building binaries..."
export PATH=$PATH:/usr/local/go/bin
go mod tidy
go build -o /opt/smm/bin/bot    ./cmd/bot
go build -o /opt/smm/bin/worker ./cmd/worker
go build -o /opt/smm/bin/server ./cmd/server
chown smm:smm /opt/smm/bin/*
echo "      Binaries OK"

# 3. Run pending migrations (idempotent — IF NOT EXISTS guards)
echo "[3/4] Running migrations..."
sudo -u postgres psql -d smm < internal/db/migrate_v2.sql 2>&1 | grep -v "^$" || true
# Grant in case new tables were added
sudo -u postgres psql -d smm -c "
  GRANT ALL PRIVILEGES ON ALL TABLES    IN SCHEMA public TO smm;
  GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO smm;
" > /dev/null
echo "      Migrations OK"

# 4. Restart services
echo "[4/4] Restarting services..."
systemctl restart smm-bot smm-worker smm-server
sleep 2
systemctl is-active --quiet smm-bot    && echo "      smm-bot     ✓" || echo "      smm-bot     FAILED"
systemctl is-active --quiet smm-worker && echo "      smm-worker  ✓" || echo "      smm-worker  FAILED"
systemctl is-active --quiet smm-server && echo "      smm-server  ✓" || echo "      smm-server  FAILED"

echo ""
echo "=== Update complete ==="
echo "Logs: journalctl -u smm-bot -u smm-worker -u smm-server -f"
