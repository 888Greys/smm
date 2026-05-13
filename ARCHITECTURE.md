# SMM Panel — System Architecture

## Overview

A Telegram-first SMM reseller platform. Clients order social media growth packages via a bot, pay via M-Pesa, and the system automatically routes orders to SMMWiz who delivers the actual followers/likes/views to the client's social media profile.

---

## The Full Flow: Client to Real Followers

```
CLIENT                   TELEGRAM BOT              YOUR SYSTEM              SMMWIZ              SOCIAL PLATFORM
  │                           │                         │                      │                      │
  │  /start or /menu          │                         │                      │                      │
  │──────────────────────────▶│                         │                      │                      │
  │                           │                         │                      │                      │
  │  Browse packages          │                         │                      │                      │
  │◀──────────────────────────│                         │                      │                      │
  │  [TikTok Starter KES1500] │                         │                      │                      │
  │  [IG Business Boost 1500] │                         │                      │                      │
  │  [YouTube Kickstart 1500] │                         │                      │                      │
  │  [Follower Booster  600 ] │                         │                      │                      │
  │                           │                         │                      │                      │
  │  Tap package              │                         │                      │                      │
  │──────────────────────────▶│                         │                      │                      │
  │                           │                         │                      │                      │
  │  "Paste your profile link"│                         │                      │                      │
  │◀──────────────────────────│                         │                      │                      │
  │                           │                         │                      │                      │
  │  https://instagram.com/.. │                         │                      │                      │
  │──────────────────────────▶│                         │                      │                      │
  │                           │  CreatePendingOrder()   │                      │                      │
  │                           │────────────────────────▶│                      │                      │
  │                           │                         │  INSERT orders       │                      │
  │                           │                         │  INSERT transactions │                      │
  │                           │                         │  status = pending    │                      │
  │                           │                         │                      │                      │
  │  "Send KES 1500 via       │                         │                      │                      │
  │   M-Pesa. Share ref code" │                         │                      │                      │
  │◀──────────────────────────│                         │                      │                      │
  │                           │                         │                      │                      │
  │  [ADMIN notified]         │                         │                      │                      │
  │                     PARTNER sees:                   │                      │                      │
  │                     "Order #42 — IG Boost           │                      │                      │
  │                      KES 1500 — link: .."           │                      │                      │
  │                     [Approve] [Reject]              │                      │                      │
  │                           │                         │                      │                      │
  │  Client sends M-Pesa ─────────────────────────────▶ Partner confirms      │                      │
  │                           │                         │                      │                      │
  │                     Partner taps [Approve]          │                      │                      │
  │                           │  ConfirmTransaction()   │                      │                      │
  │                           │────────────────────────▶│                      │                      │
  │                           │                         │  transactions.confirmed = true               │
  │                           │                         │                      │                      │
  │                           │  fulfillOrder() ───────▶│                      │                      │
  │                           │                         │  POST /api/v2        │                      │
  │                           │                         │  action=add          │                      │
  │                           │                         │  service=X           │                      │
  │                           │                         │  link=...            │                      │
  │                           │                         │  quantity=1500 ─────▶│                      │
  │                           │                         │                      │  Queues delivery     │
  │                           │                         │◀─────────────────────│                      │
  │                           │                         │  {order: 9921}       │                      │
  │                           │                         │                      │                      │
  │                           │                         │  wiz_order_ids saved │                      │
  │                           │                         │  status = processing │                      │
  │                           │                         │                      │  Delivers followers ▶│
  │                           │                         │                      │  to client's profile │
  │                           │                         │                      │                      │
  │  [STATUS POLLER runs every 30min]                   │                      │                      │
  │                           │                         │  GET action=status   │                      │
  │                           │                         │  order=9921 ────────▶│                      │
  │                           │                         │◀─────────────────────│                      │
  │                           │                         │  {status:Completed}  │                      │
  │                           │                         │  UPDATE orders       │                      │
  │                           │                         │  status = completed  │                      │
  │                           │                         │                      │                      │
  │  "Your order is complete!"│                         │                      │                      │
  │◀──────────────────────────│                         │                      │                      │
```

---

## System Components

### 1. Telegram Bot (`cmd/bot`)
The only client-facing interface. Handles all user interactions.

| Responsibility | Detail |
|---|---|
| Package menu | Inline keyboard with 4 packages |
| Link collection | Prompts client for profile/post URL, validates format |
| Order creation | Writes pending order to DB |
| Payment instruction | Tells client to send M-Pesa and the amount |
| Admin notification | Pushes order card to partner with Approve/Reject buttons |
| Fulfillment trigger | On approval, calls SMMWiz API to place sub-orders |
| Status updates | Notifies client when order completes or fails |

**Session state** is held in-memory per chat (upgrade to Redis for scale).

---

### 2. SMMWiz API Client (`internal/smmwiz`)
A thin Go HTTP client wrapping the SMMWiz REST API at `https://smmwiz.com/api/v2`.

| Method | Action | Used for |
|---|---|---|
| `AddOrder` | `add` | Place follower/view/like orders |
| `GetStatus` | `status` | Poll single order progress |
| `GetMultiStatus` | `status` | Poll all open orders in one call |
| `GetServices` | `services` | Discover available services + IDs |
| `Refill` | `refill` | Trigger 30-day refill for Follower Booster |
| `GetRefillStatus` | `refill_status` | Check if refill completed |
| `CancelOrders` | `cancel` | Cancel on client request or failure |
| `GetBalance` | `balance` | Admin wallet check |

All requests are `POST application/x-www-form-urlencoded`. All responses are JSON.

---

### 3. PostgreSQL Database (`internal/db`)

```
┌──────────┐       ┌──────────────┐       ┌────────────────┐
│ clients  │──────▶│    orders    │──────▶│  transactions  │
│          │       │              │       │                │
│ id       │       │ id           │       │ id             │
│ telegram_│       │ client_id    │       │ order_id       │
│   _id    │       │ package_id   │       │ amount_kes     │
│ username │       │ profile_link │       │ mpesa_ref      │
│ phone    │       │ total_kes    │       │ confirmed      │
└──────────┘       │ status       │       │ confirmed_by   │
                   │ wiz_order_ids│       │ confirmed_at   │
                   └──────────────┘       └────────────────┘
                          │
                          ▼
                   ┌────────────────┐
                   │ refill_records │
                   │                │
                   │ order_id       │
                   │ wiz_order_id   │
                   │ wiz_refill_id  │
                   │ status         │
                   └────────────────┘
```

**Order status lifecycle:**
```
pending → processing → completed
                    ↘ partial
                    ↘ failed → [manual retry or refund]
```

---

### 4. Background Worker (`cmd/worker`)
Runs on a ticker independently of the bot process.

```
Every 30 minutes:
  1. Fetch all orders WHERE status = 'processing'
  2. Batch-call GetMultiStatus on their wiz_order_ids
  3. For each:
     - Completed → update DB, notify client via bot
     - Failed    → update DB, alert admin, optionally cancel
     - Partial   → log remains, continue polling

Every 24 hours:
  1. Fetch Follower Booster orders WHERE status = 'completed'
     AND created_at > 30 days ago
     AND no refill triggered yet
  2. Call Refill() on each wiz_order_id
  3. Save refill_record

Every morning:
  1. Call GetBalance()
  2. If balance < threshold → send low-balance alert to admin
```

---

### 5. Payment Flow (Manual → Automated path)

**Phase 1 (Now): Manual M-Pesa Confirmation**
```
Client pays M-Pesa → tells partner via DM → partner taps [Approve] in bot
```
No webhook, no integration needed. Partner is the human payment gate.

**Phase 2 (Month 2): MegaPay Webhook**
```
Client pays M-Pesa → MegaPay fires webhook → system auto-confirms → order placed instantly
```
Endpoint: `POST /webhook/megapay`
Validates HMAC signature, matches amount + order ID, calls `ConfirmTransaction()`.

---

## Service Packages & SMMWiz Mapping

Each retail package splits into N SMMWiz sub-orders placed in sequence:

| Package | KES | Sub-orders placed |
|---|---|---|
| TikTok Viral Starter | 1,500 | Followers (2k) + Views (5k) + Likes (200) |
| IG Business Boost | 1,500 | Followers (1.5k) + Likes (300) |
| YouTube Kickstart | 1,500 | Subscribers (300) + Views (1k) |
| Follower Booster | 600 | Followers (1k) — refill-eligible |

**Drip-feed** is set at the component level via `Runs` + `Interval` fields in `PackageComponent`. Example: 2,000 TikTok followers over 10 days = `runs: 10, interval: 1440` (1440 min = 24h).

---

## Data & Money Flow

```
Client KES 1,500
    │
    ▼
M-Pesa (MegaPay)
    │
    ▼
Your wallet (net: KES 1,500 minus ~1% M-Pesa fee)
    │
    ▼
SMMWiz wallet debit (~KES 300–500 wholesale)
    │
    ▼
Margin retained: ~KES 1,000–1,200 per order
```

Your SMMWiz wallet is pre-funded (top up manually or via auto-topup when balance drops below threshold). **Orders fail silently if the wallet is empty** — the balance alert worker prevents this.

---

## Deployment

```
┌──────────────────────────────────────┐
│             VPS / Cloud VM           │
│                                      │
│  ┌─────────────┐  ┌───────────────┐  │
│  │  bot binary │  │worker binary  │  │
│  │  (always on)│  │(cron/ticker)  │  │
│  └──────┬──────┘  └───────┬───────┘  │
│         │                 │          │
│         └────────┬────────┘          │
│                  ▼                   │
│           ┌────────────┐             │
│           │ PostgreSQL │             │
│           └────────────┘             │
│                                      │
│  Optional: nginx reverse proxy       │
│  for MegaPay webhook endpoint        │
└──────────────────────────────────────┘
         │                 │
         ▼                 ▼
   Telegram API       SMMWiz API
   (polling)         (HTTP POST)
```

Single VPS (2GB RAM, 1 vCPU) is sufficient for this volume. Both binaries compile to single static executables — no runtime dependencies.

---

## What You Do vs What Runs Automatically

| Action | Who handles it |
|---|---|
| Client taps package, submits link | Bot (automated) |
| Client receives payment instructions | Bot (automated) |
| Partner confirms M-Pesa payment | **Partner — one tap** |
| SMMWiz order placement | Bot (automated, on approval) |
| Order status polling | Worker (automated, every 30min) |
| Client notified on completion | Worker (automated) |
| 30-day refill trigger | Worker (automated) |
| Low wallet balance alert | Worker (automated) |
| SMMWiz wallet top-up | **You — manual** |

After Phase 2 (MegaPay webhook), the partner's one tap is also removed.

---

## Environment Variables

```
TELEGRAM_BOT_TOKEN      — from @BotFather
MTP_API_KEY             — from morethanpanel.com dashboard (API key)
ADMIN_TELEGRAM_IDS      — your ID + partner's ID, comma-separated
DATABASE_URL            — postgres://user:pass@host:5432/smm
MEGAPAY_WEBHOOK_SECRET  — for Phase 2 only
BALANCE_ALERT_THRESHOLD — morethanpanel wallet floor (e.g. 10.00 USD)
```
