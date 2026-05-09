CREATE TABLE clients (
    id                  BIGSERIAL PRIMARY KEY,
    telegram_id         BIGINT UNIQUE NOT NULL,
    username            TEXT,
    phone               TEXT,
    referral_code       TEXT UNIQUE,
    credit_balance_kes  INT NOT NULL DEFAULT 0,
    referred_by         BIGINT REFERENCES clients(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id            BIGSERIAL PRIMARY KEY,
    client_id     BIGINT NOT NULL REFERENCES clients(id),
    package_id    TEXT NOT NULL,
    profile_link  TEXT NOT NULL,
    total_kes     INT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    wiz_order_ids BIGINT[],
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    id                     BIGSERIAL PRIMARY KEY,
    order_id               BIGINT NOT NULL REFERENCES orders(id),
    amount_kes             INT NOT NULL,
    phone                  TEXT,
    mpesa_ref              TEXT,
    stk_request_id         TEXT,
    confirmed              BOOLEAN NOT NULL DEFAULT FALSE,
    confirmed_by           BIGINT,
    confirmed_at           TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refill_records (
    id            BIGSERIAL PRIMARY KEY,
    order_id      BIGINT NOT NULL REFERENCES orders(id),
    wiz_order_id  BIGINT NOT NULL,
    wiz_refill_id BIGINT,
    status        TEXT NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON orders(client_id);
CREATE INDEX ON orders(status);
CREATE INDEX ON transactions(order_id);
CREATE INDEX ON transactions(confirmed);
CREATE INDEX ON transactions(stk_request_id);
CREATE INDEX ON refill_records(order_id);
CREATE INDEX ON refill_records(status);

CREATE TABLE referrals (
    id          BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES clients(id),
    referred_id BIGINT NOT NULL UNIQUE REFERENCES clients(id),
    order_id    BIGINT REFERENCES orders(id),
    credit_kes  INT NOT NULL DEFAULT 50,
    paid        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON referrals(referrer_id);
