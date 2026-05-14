# VectorBoost SMM Panel

Social media growth platform for Kenyan market. M-Pesa payments via MegaPay, order fulfillment via morethanpanel.com, orders placed through website or Telegram bot.

---

## Architecture

```
innbucks.org (Vercel frontend)
    └── calls → api.innbucks.org (Go API, port 8005, nginx reverse proxy)
                    ├── MegaPay (M-Pesa STK push)
                    ├── morethanpanel.com (SMM order fulfillment)
                    └── PostgreSQL (orders, bot_sessions, transactions)

@pompomputrin888pom_bot (Telegram ordering bot)
    └── same backend: morethanpanel.com + MegaPay + PostgreSQL

@POM888POM_bot (Admin notifications bot)
```

---

## Server

**Host:** `vmi3255937` — IP `79.143.178.165`  
**SSH:** `ssh root@79.143.178.165`  
**Source code:** `/root/smm` (git pull here)  
**Compiled binaries:** `/opt/smm/bin/` (systemd runs these)  
**Environment file:** `/opt/smm/.env` ← this is what systemd reads, NOT `/root/smm/.env`

---

## Services

All three are managed by systemd and start automatically on reboot.

| Service | Description | Binary |
|---------|-------------|--------|
| `smm-bot` | Telegram ordering bot | `/opt/smm/bin/bot` |
| `smm-server` | REST API + MegaPay webhook (`:8005`) | `/opt/smm/bin/server` |
| `smm-worker` | Background payment poller + order fulfiller | `/opt/smm/bin/worker` |

### Status & logs

```bash
# Status of all three
systemctl status smm-bot smm-server smm-worker --no-pager

# Live logs
journalctl -u smm-bot -f
journalctl -u smm-server -f
journalctl -u smm-worker -f

# Last 30 lines
journalctl -u smm-worker --no-pager | tail -30
```

---

## Deploying updates (CORRECT method)

> **Important:** `go build ./...` builds binaries to a temp location. Systemd runs from `/opt/smm/bin/`. You MUST use `-o` to build directly into that path, otherwise the old binary keeps running.

```bash
cd ~/smm && git fetch origin && git reset --hard origin/main \
  && go build -o /opt/smm/bin/server ./cmd/server \
  && go build -o /opt/smm/bin/bot ./cmd/bot \
  && go build -o /opt/smm/bin/worker ./cmd/worker \
  && systemctl restart smm-server smm-bot smm-worker
```

If `go mod tidy` is needed (you'll see a warning):
```bash
go mod tidy && go build -o /opt/smm/bin/server ./cmd/server \
  && go build -o /opt/smm/bin/bot ./cmd/bot \
  && go build -o /opt/smm/bin/worker ./cmd/worker \
  && systemctl restart smm-server smm-bot smm-worker
```

Partial restarts (when only one service changed):
```bash
# API only
go build -o /opt/smm/bin/server ./cmd/server && systemctl restart smm-server

# Bot only
go build -o /opt/smm/bin/bot ./cmd/bot && systemctl restart smm-bot

# Worker only
go build -o /opt/smm/bin/worker ./cmd/worker && systemctl restart smm-worker
```

---

## Environment variables

Stored in **`/opt/smm/.env`** — this is what systemd reads at startup.

> **Do NOT edit `/root/smm/.env`** — systemd ignores it. Always edit `/opt/smm/.env`.

To update a value:
```bash
sed -i 's/KEY_NAME=.*/KEY_NAME=newvalue/' /opt/smm/.env
systemctl restart smm-server smm-bot smm-worker
```

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `MEGAPAY_API_KEY` | MegaPay payment gateway key |
| `MEGAPAY_EMAIL` | MegaPay account email (toxicgreys001@gmail.com) |
| `MEGAPAY_WEBHOOK_SECRET` | Webhook HMAC secret |
| `MTP_API_KEY` | morethanpanel.com API key |
| `TELEGRAM_BOT_TOKEN` | Main bot token (messages clients) |
| `ADMIN_BOT_TOKEN` | Admin bot token (messages admin) |
| `ADMIN_CHAT_ID` | Admin Telegram chat ID |
| `ADMIN_TELEGRAM_IDS` | Comma-separated admin Telegram IDs |
| `BALANCE_ALERT_THRESHOLD` | Alert when morethanpanel balance drops below this (USD) |
| `FRONTEND_ORIGIN` | Allowed CORS origin (`https://innbucks.org`) |
| `SOCIAL_PROOF_CHANNEL_ID` | Optional Telegram channel for order completion posts |

---

## Database

**Engine:** PostgreSQL  
**DB name:** `smm` | **User:** `smm`

```bash
# Connect
source /opt/smm/.env && psql "$DATABASE_URL"

# View recent orders
source /opt/smm/.env && psql "$DATABASE_URL" -c \
  "SELECT id, package_id, profile_link, status, wiz_order_ids FROM orders ORDER BY id DESC LIMIT 10;"

# Check a specific order
source /opt/smm/.env && psql "$DATABASE_URL" -c \
  "SELECT id, profile_link, status, wiz_order_ids FROM orders WHERE id = 20;"
```

### Granting schema permissions (PostgreSQL 15+ gotcha)

If a new migration fails with `permission denied for schema public`:
```bash
sudo -u postgres psql -d smm -c "GRANT CREATE ON SCHEMA public TO smm;"
```
Then retry the migration or restart the service.

### Tables

| Table | Purpose |
|-------|---------|
| `clients` | Telegram users, referral codes, credit balances |
| `orders` | Every order placed (pending → processing → completed) |
| `transactions` | M-Pesa STK push records |
| `bot_sessions` | Persisted Telegram bot session state (survives restarts) |
| `refill_records` | 30-day refill tracking |
| `referrals` | Referral credit records |

---

## Nginx

Config lives in `/etc/nginx/sites-enabled/`.

| Domain | Routes to |
|--------|-----------|
| `api.innbucks.org` | `127.0.0.1:8005` (Go API server) |
| `innbucks.org` | Vercel (DNS at registrar points to Vercel, NOT this VPS) |
| `bot.innbucks.org` | `127.0.0.1:3000` |
| `status.innbucks.org` | `127.0.0.1:3001` |

```bash
nginx -t && systemctl reload nginx
```

---

## Frontend (Vercel)

- Repo: `web/` directory (Next.js)
- Deployed automatically on every push to `main`
- API URL: set `NEXT_PUBLIC_API_URL=https://api.innbucks.org` in Vercel project env vars
- DNS for `innbucks.org` points to Vercel, **not** the VPS

To force a Vercel redeploy without code changes: push any trivial change to `web/`.

---

## Packages / Catalog

Defined in `internal/bot/packages.go`. Provider is **morethanpanel.com**.

| ID | Name | Price (KES) | Platform |
|----|------|-------------|----------|
| `tiktok_flex` | TikTok Quick-Start | 500 | TikTok |
| `tiktok_starter` | TikTok Starter | 1,000 | TikTok |
| `tiktok_viral_starter` | TikTok Viral Starter | 1,500 | TikTok |
| `ig_quick_start` | IG Quick-Start | 500 | Instagram |
| `ig_business_boost` | IG Business Boost | 800 | Instagram |
| `follower_booster` | Follower Booster | 600 | Instagram |
| `ig_celebrity_pack` | IG Celebrity Pack | 2,500 | Instagram |
| `youtube_kickstart` | YouTube Kickstart | 1,500 | YouTube |
| `viral_creator_combo` | Viral Creator Combo | 2,500 | TikTok combo |

Service IDs (morethanpanel.com):
- TikTok Followers: `5760` — $2.44/1k, 30-day refill
- TikTok Views: `9121` — $0.04/1k, 30-day refill
- TikTok Likes: `2699` — $0.32/1k, 30-day refill
- IG Followers: `5440` — $0.35/1k, 30-day refill
- IG Likes: `2916` — $0.10/1k, 30-day refill
- YT Subscribers: `7494` — $0.70/1k
- YT Views: `6003` — $0.41/1k, lifetime guarantee

To add/change a package: edit `internal/bot/packages.go` then redeploy.

---

## Payment flow (website)

1. User picks a package on `innbucks.org`
2. Frontend calls `POST /api/orders` with package, profile link, phone
3. Server creates a pending order in DB, calls MegaPay STK push
4. User gets M-Pesa PIN prompt on phone
5. Worker polls MegaPay every 10s for payment confirmation
6. On confirmation, worker submits order to morethanpanel.com
7. Worker polls morethanpanel every 2 minutes until order completes
8. Telegram notification sent to client when done

## Payment flow (Telegram bot)

1. User messages `@pompomputrin888pom_bot`, picks a package
2. Bot scans profile (verifies it's public), shows profile card
3. User confirms → enters M-Pesa number
4. Same STK push → worker polling → fulfillment flow as above
5. Bot sessions are persisted to DB — users survive bot restarts mid-flow

---

## Verifying an order end-to-end

```bash
# 1. Check worker logs for payment + fulfillment
journalctl -u smm-worker --no-pager | grep "order 20"

# 2. Check DB for order status and morethanpanel IDs
source /opt/smm/.env && psql "$DATABASE_URL" -c \
  "SELECT id, status, wiz_order_ids FROM orders WHERE id = 20;"

# 3. Check morethanpanel dashboard
# → morethanpanel.com/orders → search for the wiz_order_id from step 2
```

Order status flow: `pending` → `processing` → `completed` (or `failed`/`cancelled`)

---

## Known issues & fixes

### 1. Bot sessions lost on restart
**Fix:** Sessions are now persisted to `bot_sessions` table in PostgreSQL. Users mid-flow survive restarts automatically.

### 2. Website "Network error" on payment
**Cause:** Go HTTP server `WriteTimeout` was 15s but MegaPay STK can take up to 30s.  
**Fix:** `WriteTimeout` raised to 35s in `cmd/server/main.go`.

### 3. Website showing empty packages
**Cause:** Vercel ISR cached the page during a server restart (API returned empty).  
**Fix:** Frontend now always does a client-side refresh on load regardless of SSR result.

### 4. Bot conflict error (`Conflict: terminated by other getUpdates`)
**Cause:** Two bot instances running — one via PM2, one via systemd.  
**Fix:** `pm2 delete bot` then `systemctl restart smm-bot`. Never run the bot via PM2 again; systemd manages it.

### 5. `go build ./...` doesn't update the running binary
**Cause:** Systemd runs from `/opt/smm/bin/`, not from the build cache.  
**Fix:** Always build with `-o /opt/smm/bin/<name>` (see deploy command above).

### 6. `permission denied for schema public` on DB migration
**Cause:** PostgreSQL 15+ removed default CREATE privileges on public schema.  
**Fix:** `sudo -u postgres psql -d smm -c "GRANT CREATE ON SCHEMA public TO smm;"`

### 7. Service fails with `required env var X is not set`
**Cause:** Editing `/root/smm/.env` instead of `/opt/smm/.env`, or binary is outdated.  
**Fix:** Edit `/opt/smm/.env`, then rebuild + restart. If binary is old, run full deploy command.

### 8. morethanpanel orders not appearing in dashboard
**Cause:** Wrong API key in `.env` — orders going to a different account.  
**Fix:** Generate new API key on morethanpanel.com → API page, update with:  
`sed -i 's/MTP_API_KEY=.*/MTP_API_KEY=newkey/' /opt/smm/.env && systemctl restart smm-server smm-worker`

### 9. `SMMWIZ_API_KEY is not set` error
**Cause:** Running an old compiled binary that still references the old provider variable.  
**Fix:** Rebuild all binaries with the deploy command above.
