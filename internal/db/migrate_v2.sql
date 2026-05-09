-- Migration v2: add referral system
-- Run on server: sudo -u postgres psql -d smm < /opt/smm/repo/internal/db/migrate_v2.sql

ALTER TABLE clients ADD COLUMN IF NOT EXISTS referral_code       TEXT UNIQUE;
ALTER TABLE clients ADD COLUMN IF NOT EXISTS credit_balance_kes  INT NOT NULL DEFAULT 0;
ALTER TABLE clients ADD COLUMN IF NOT EXISTS referred_by         BIGINT REFERENCES clients(id);

CREATE TABLE IF NOT EXISTS referrals (
    id          BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES clients(id),
    referred_id BIGINT NOT NULL UNIQUE REFERENCES clients(id),
    order_id    BIGINT REFERENCES orders(id),
    credit_kes  INT NOT NULL DEFAULT 50,
    paid        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS referrals_referrer_idx ON referrals(referrer_id);
