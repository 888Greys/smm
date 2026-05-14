# VectorBoost SMM Panel

Social media growth platform for Kenyan market. M-Pesa payments via MegaPay, order fulfillment via morethanpanel.com, orders placed through website or Telegram bot.

---

## Architecture

```
innbucks.org (Vercel frontend)
    └── calls → api.innbucks.org (Go API, port 8005, nginx reverse proxy)
                    ├── MegaPay (M-Pesa STK push)
                    ├── morethanpanel.com (SMM order fulfillment)
                    └── PostgreSQL (orders, sessions, transactions)

@pompomputrin888pom_bot (Telegram ordering bot)
    └── same backend: morethanpanel.com + MegaPay + PostgreSQL

@POM888POM_bot (Admin notifications bot)
```

---

## Server

**Host:** `vmi3255937` — IP `79.143.178.165`  
**SSH:** `ssh root@79.143.178.165`  
**Project directory:** `/root/smm`

---

## Services

All three are managed by systemd and start automatically on reboot.

| Service | Description | Binary |
|---------|-------------|--------|
| `smm-bot` | Telegram ordering bot | `/opt/smm/bin/bot` |
| `smm-server` | REST API + MegaPay webhook (`:8005`) | `/opt/smm/bin/server` |
| `smm-worker` | Background order poller/fulfiller | `/opt/smm/bin/worker` |

### Common commands

```bash
# Status
systemctl status smm-bot smm-server smm-worker

# Restart all
systemctl restart smm-bot smm-server smm-worker

# Logs (live)
journalctl -u smm-bot -f
journalctl -u smm-server -f
journalctl -u smm-worker -f
```

---

## Deploying updates

```bash
cd ~/smm
git fetch origin && git reset --hard origin/main
go mod tidy
go build ./...
systemctl restart smm-bot smm-server smm-worker
```

> If only the API changed: `systemctl restart smm-server` is enough.  
> If only the bot changed: `systemctl restart smm-bot` is enough.

---

## Nginx

Config lives in `/etc/nginx/sites-enabled/`.

| Domain | Routes to |
|--------|-----------|
| `api.innbucks.org` | `127.0.0.1:8005` (Go API server) |
| `innbucks.org` | Vercel (DNS points to Vercel, not this VPS) |
| `bot.innbucks.org` | `127.0.0.1:3000` |
| `status.innbucks.org` | `127.0.0.1:3001` |

Reload nginx after config changes:
```bash
nginx -t && systemctl reload nginx
```

---

## Environment variables

Stored in `/root/smm/.env` (loaded by `godotenv` at startup).

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `MEGAPAY_API_KEY` | MegaPay payment gateway key |
| `MEGAPAY_EMAIL` | MegaPay account email |
| `MEGAPAY_WEBHOOK_SECRET` | Webhook HMAC secret |
| `MTP_API_KEY` | morethanpanel.com API key |
| `TELEGRAM_BOT_TOKEN` | Main bot token (messages clients) |
| `ADMIN_BOT_TOKEN` | Admin bot token (messages admin) |
| `ADMIN_CHAT_ID` | Admin Telegram chat ID |
| `FRONTEND_ORIGIN` | Allowed CORS origin (e.g. `https://innbucks.org`) |

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

---

## Payment flow (website)

1. User picks a package on `innbucks.org`
2. Frontend calls `POST /api/orders` with package, profile link, phone
3. Server creates a pending order in DB, calls MegaPay STK push
4. User gets M-Pesa PIN prompt on phone
5. MegaPay calls `POST /webhook/megapay` on payment success
6. Server auto-fulfills the order via morethanpanel.com
7. Telegram notification sent to client and admin

## Payment flow (Telegram bot)

1. User messages `@pompomputrin888pom_bot`, picks a package
2. Bot collects profile link and M-Pesa number
3. Same STK push → webhook → fulfillment flow as above

---

## Frontend (Vercel)

Repo: `web/` directory (Next.js).  
Deployed on Vercel — DNS for `innbucks.org` points to Vercel, **not** the VPS.  
API URL configured via `NEXT_PUBLIC_API_URL=https://api.innbucks.org` in Vercel env vars.

---

## Known fixes applied

- **WriteTimeout raised to 35s** (`cmd/server/main.go`) — MegaPay STK push can take up to 30s; the original 15s timeout caused the browser to see a network error before the response arrived.
- **Provider switched from SMMWiz → morethanpanel.com** — better pricing and reliability.
