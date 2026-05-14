package db

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/aapom/smm/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, connString string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS bot_sessions (
			telegram_id   BIGINT PRIMARY KEY,
			step          TEXT NOT NULL DEFAULT '',
			package_id    TEXT NOT NULL DEFAULT '',
			profile_link  TEXT NOT NULL DEFAULT '',
			referral_code TEXT NOT NULL DEFAULT '',
			scan_msg_id   INT  NOT NULL DEFAULT 0,
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return nil, fmt.Errorf("migrate bot_sessions: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

// UpsertClient ensures a client row exists. Returns true if the client is new.
func (s *Store) UpsertClient(ctx context.Context, telegramID int64) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO clients (telegram_id) VALUES ($1)
		ON CONFLICT (telegram_id) DO NOTHING
	`, telegramID)
	return tag.RowsAffected() == 1, err
}

func (s *Store) CreatePendingOrder(ctx context.Context, clientTelegramID int64, packageID, link string, amountKES int, referralCode string) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var clientID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO clients (telegram_id) VALUES ($1)
		ON CONFLICT (telegram_id) DO UPDATE SET telegram_id = EXCLUDED.telegram_id
		RETURNING id
	`, clientTelegramID).Scan(&clientID)
	if err != nil {
		return 0, fmt.Errorf("upsert client: %w", err)
	}

	// Wire up referral if provided and not already set
	if referralCode != "" {
		tx.Exec(ctx, `
			UPDATE clients
			SET referred_by = (SELECT id FROM clients WHERE referral_code = $1)
			WHERE id = $2 AND referred_by IS NULL
		`, referralCode, clientID)
	}

	var orderID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (client_id, package_id, profile_link, total_kes, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id
	`, clientID, packageID, link, amountKES).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("insert order: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (order_id, amount_kes) VALUES ($1, $2)
	`, orderID, amountKES)
	if err != nil {
		return 0, fmt.Errorf("insert transaction: %w", err)
	}

	return orderID, tx.Commit(ctx)
}

func (s *Store) SaveSTKRequest(ctx context.Context, orderID int64, phone, stkRequestID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE transactions SET phone = $1, stk_request_id = $2 WHERE order_id = $3
	`, phone, stkRequestID, orderID)
	return err
}

func (s *Store) ConfirmTransaction(ctx context.Context, orderID, confirmedBy int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE transactions
		SET confirmed = true, confirmed_by = $1, confirmed_at = NOW()
		WHERE order_id = $2
	`, confirmedBy, orderID)
	return err
}

func (s *Store) CancelOrder(ctx context.Context, orderID int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE orders SET status = 'cancelled', updated_at = NOW() WHERE id = $1
	`, orderID)
	return err
}

func (s *Store) GetOrder(ctx context.Context, orderID int64) (*models.Order, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, client_id, package_id, profile_link, total_kes, status, wiz_order_ids, created_at, updated_at
		FROM orders WHERE id = $1
	`, orderID)

	o := &models.Order{}
	err := row.Scan(&o.ID, &o.ClientID, &o.PackageID, &o.ProfileLink, &o.TotalKES,
		&o.Status, &o.WizOrderIDs, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return o, nil
}

func (s *Store) UpdateOrderStatus(ctx context.Context, orderID int64, status models.OrderStatus, wizIDs []int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE orders SET status = $1, wiz_order_ids = $2, updated_at = NOW() WHERE id = $3
	`, string(status), wizIDs, orderID)
	return err
}

func (s *Store) SaveRefill(ctx context.Context, orderID, wizOrderID, wizRefillID int64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO refill_records (order_id, wiz_order_id, wiz_refill_id, status)
		VALUES ($1, $2, $3, 'pending')
	`, orderID, wizOrderID, wizRefillID)
	return err
}

func (s *Store) GetProcessingOrders(ctx context.Context) ([]*models.Order, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, client_id, package_id, profile_link, total_kes, status, wiz_order_ids, created_at, updated_at
		FROM orders WHERE status = 'processing' AND wiz_order_ids IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{}
		if err := rows.Scan(&o.ID, &o.ClientID, &o.PackageID, &o.ProfileLink, &o.TotalKES,
			&o.Status, &o.WizOrderIDs, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (s *Store) GetClientTelegramID(ctx context.Context, orderID int64) (int64, error) {
	var tgID int64
	err := s.pool.QueryRow(ctx, `
		SELECT c.telegram_id FROM clients c
		JOIN orders o ON o.client_id = c.id
		WHERE o.id = $1
	`, orderID).Scan(&tgID)
	return tgID, err
}

// PendingSTKTransaction holds data needed to poll a payment
type PendingSTKTransaction struct {
	OrderID      int64
	STKRequestID string
	AmountKES    int
}

func (s *Store) GetPendingSTKTransactions(ctx context.Context) ([]PendingSTKTransaction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.order_id, t.stk_request_id, t.amount_kes
		FROM transactions t
		JOIN orders o ON o.id = t.order_id
		WHERE t.confirmed = false
		  AND t.stk_request_id IS NOT NULL
		  AND o.status = 'pending'
		  AND t.created_at > NOW() - INTERVAL '30 minutes'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []PendingSTKTransaction
	for rows.Next() {
		var t PendingSTKTransaction
		if err := rows.Scan(&t.OrderID, &t.STKRequestID, &t.AmountKES); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

func (s *Store) GetRefillableOrders(ctx context.Context, packageIDs []string) ([]*models.Order, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT o.id, o.client_id, o.package_id, o.profile_link, o.total_kes,
		       o.status, o.wiz_order_ids, o.created_at, o.updated_at
		FROM orders o
		WHERE o.package_id = ANY($1::text[])
		  AND o.status = 'completed'
		  AND o.created_at <= NOW() - INTERVAL '30 days'
		  AND NOT EXISTS (
		    SELECT 1 FROM refill_records r WHERE r.order_id = o.id
		  )
	`, packageIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{}
		if err := rows.Scan(&o.ID, &o.ClientID, &o.PackageID, &o.ProfileLink, &o.TotalKES,
			&o.Status, &o.WizOrderIDs, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// GetOrCreateReferralCode returns the referral code for the given telegram user,
// generating one if needed. The client must exist (call UpsertClient first).
func (s *Store) GetOrCreateReferralCode(ctx context.Context, telegramID int64) (string, error) {
	var code *string
	err := s.pool.QueryRow(ctx, `
		SELECT referral_code FROM clients WHERE telegram_id = $1
	`, telegramID).Scan(&code)
	if err != nil {
		return "", fmt.Errorf("get referral code: %w", err)
	}
	if code != nil && *code != "" {
		return *code, nil
	}

	// Generate and save
	newCode := generateReferralCode()
	_, err = s.pool.Exec(ctx, `
		UPDATE clients SET referral_code = $1 WHERE telegram_id = $2
	`, newCode, telegramID)
	if err != nil {
		return "", fmt.Errorf("save referral code: %w", err)
	}
	return newCode, nil
}

// GetCreditBalance returns the KES credit balance for a telegram user.
func (s *Store) GetCreditBalance(ctx context.Context, telegramID int64) (int, error) {
	var bal int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(credit_balance_kes, 0) FROM clients WHERE telegram_id = $1
	`, telegramID).Scan(&bal)
	return bal, err
}

// AwardReferralCredit pays KES 50 to the referrer of the order's client (once per referred user).
// Returns the referrer's Telegram ID so the caller can notify them, or 0 if no referral.
func (s *Store) AwardReferralCredit(ctx context.Context, orderID int64) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Get referred client and their referrer
	var referredClientID, referrerClientID int64
	err = tx.QueryRow(ctx, `
		SELECT c.id, c.referred_by
		FROM orders o JOIN clients c ON c.id = o.client_id
		WHERE o.id = $1 AND c.referred_by IS NOT NULL
	`, orderID).Scan(&referredClientID, &referrerClientID)
	if err != nil {
		return 0, nil // no referral, not an error
	}

	// Only pay once per referred user
	var alreadyPaid bool
	tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM referrals WHERE referred_id = $1 AND paid = true)
	`, referredClientID).Scan(&alreadyPaid)
	if alreadyPaid {
		return 0, nil
	}

	// Credit referrer
	if _, err = tx.Exec(ctx, `
		UPDATE clients SET credit_balance_kes = credit_balance_kes + 50 WHERE id = $1
	`, referrerClientID); err != nil {
		return 0, err
	}

	// Record
	if _, err = tx.Exec(ctx, `
		INSERT INTO referrals (referrer_id, referred_id, order_id, credit_kes, paid)
		VALUES ($1, $2, $3, 50, true)
	`, referrerClientID, referredClientID, orderID); err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}

	// Get referrer's telegram ID for notification
	var referrerTgID int64
	s.pool.QueryRow(ctx, `SELECT telegram_id FROM clients WHERE id = $1`, referrerClientID).Scan(&referrerTgID)
	return referrerTgID, nil
}

// GetStats returns order statistics for the admin dashboard.
func (s *Store) GetStats(ctx context.Context) (*models.DailyStats, error) {
	st := &models.DailyStats{}

	rows, err := s.pool.Query(ctx, `
		SELECT o.package_id, COUNT(*), COALESCE(SUM(o.total_kes), 0)
		FROM orders o
		JOIN transactions t ON t.order_id = o.id
		WHERE t.confirmed = true
		  AND t.confirmed_at >= NOW() - INTERVAL '24 hours'
		GROUP BY o.package_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line models.PackageStatLine
		if err := rows.Scan(&line.PackageID, &line.OrderCount, &line.RevenueKES); err != nil {
			return nil, err
		}
		st.Lines = append(st.Lines, line)
	}

	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE status = 'pending'`).Scan(&st.PendingOrders)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE status = 'processing'`).Scan(&st.ProcessingOrders)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE status = 'completed'`).Scan(&st.CompletedOrders)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&st.TotalOrders)

	return st, rows.Err()
}

const referralChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func (s *Store) SaveSession(ctx context.Context, telegramID int64, step, packageID, profileLink, referralCode string, scanMsgID int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO bot_sessions (telegram_id, step, package_id, profile_link, referral_code, scan_msg_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (telegram_id) DO UPDATE SET
			step          = EXCLUDED.step,
			package_id    = EXCLUDED.package_id,
			profile_link  = EXCLUDED.profile_link,
			referral_code = EXCLUDED.referral_code,
			scan_msg_id   = EXCLUDED.scan_msg_id,
			updated_at    = NOW()
	`, telegramID, step, packageID, profileLink, referralCode, scanMsgID)
	return err
}

func (s *Store) LoadSession(ctx context.Context, telegramID int64) (step, packageID, profileLink, referralCode string, scanMsgID int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT step, package_id, profile_link, referral_code, scan_msg_id
		FROM bot_sessions
		WHERE telegram_id = $1 AND updated_at > NOW() - INTERVAL '24 hours'
	`, telegramID).Scan(&step, &packageID, &profileLink, &referralCode, &scanMsgID)
	if err != nil {
		return "", "", "", "", 0, nil // not found or expired — return empty
	}
	return
}

func generateReferralCode() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = referralChars[rand.Intn(len(referralChars))]
	}
	return string(b)
}
