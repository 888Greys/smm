#!/bin/bash
# Run once on a fresh Ubuntu/Debian VPS as root.
# Usage: bash deploy/setup.sh

set -e

echo "=== SMM Panel Setup ==="

# 1. System deps
apt update && apt install -y git golang-go postgresql nginx certbot python3-certbot-nginx

# 2. Create system user
useradd -r -m -d /opt/smm -s /bin/bash smm 2>/dev/null || true

# 3. Clone / pull repo
if [ -d /opt/smm/repo ]; then
  echo "Pulling latest..."
  git -C /opt/smm/repo pull
else
  echo "Cloning repo..."
  git clone https://github.com/888Greys/smm.git /opt/smm/repo
fi

# 4. Build binaries
mkdir -p /opt/smm/bin
cd /opt/smm/repo
go mod tidy
go build -o /opt/smm/bin/bot    ./cmd/bot
go build -o /opt/smm/bin/worker ./cmd/worker
go build -o /opt/smm/bin/server ./cmd/server
chown -R smm:smm /opt/smm/bin
echo "Binaries built."

# 5. Database
echo "Setting up PostgreSQL..."
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='smm'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE DATABASE smm;"
sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='smm'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE USER smm WITH PASSWORD 'changeme';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE smm TO smm;"
sudo -u postgres psql -d smm -f /opt/smm/repo/internal/db/schema.sql 2>/dev/null || true
echo "Database ready."

# 6. .env
if [ ! -f /opt/smm/.env ]; then
  cp /opt/smm/repo/.env.example /opt/smm/.env
  chown smm:smm /opt/smm/.env
  chmod 600 /opt/smm/.env
  echo ""
  echo ">>> FILL IN /opt/smm/.env with your keys before starting services <<<"
  echo ""
fi

# 7. Nginx
cp /opt/smm/repo/deploy/nginx/innbucks.org.conf /etc/nginx/sites-available/innbucks.org
ln -sf /etc/nginx/sites-available/innbucks.org /etc/nginx/sites-enabled/innbucks.org
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx
echo "Nginx configured."

# 8. SSL cert
echo "Getting SSL certificate for innbucks.org..."
certbot --nginx -d innbucks.org -d www.innbucks.org --non-interactive --agree-tos -m admin@innbucks.org || \
  echo "Certbot failed — run manually: certbot --nginx -d innbucks.org"

# 9. Systemd services
cp /opt/smm/repo/deploy/systemd/smm-bot.service    /etc/systemd/system/
cp /opt/smm/repo/deploy/systemd/smm-worker.service /etc/systemd/system/
cp /opt/smm/repo/deploy/systemd/smm-server.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable smm-bot smm-worker smm-server
echo ""
echo "=== Setup complete ==="
echo ""
echo "Next steps:"
echo "  1. Edit /opt/smm/.env — add your API keys"
echo "  2. systemctl start smm-bot smm-worker smm-server"
echo "  3. Check logs: journalctl -u smm-bot -f"
echo ""
