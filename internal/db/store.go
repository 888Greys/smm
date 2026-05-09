package db

import (
	"context"
	"fmt"

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
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) CreatePendingOrder(ctx context.Context, clientTelegramID int64, packageID, link string, amountKES int) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Upsert client
	var clientID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO clients (telegram_id) VALUES ($1)
		ON CONFLICT (telegram_id) DO UPDATE SET telegram_id = EXCLUDED.telegram_id
		RETURNING id
	`, clientTelegramID).Scan(&clientID)
	if err != nil {
		return 0, fmt.Errorf("upsert client: %w", err)
	}

	// Create order
	var orderID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (client_id, package_id, profile_link, total_kes, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id
	`, clientID, packageID, link, amountKES).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("insert order: %w", err)
	}

	// Create transaction record
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (order_id, amount_kes) VALUES ($1, $2)
	`, orderID, amountKES)
	if err != nil {
		return 0, fmt.Errorf("insert transaction: %w", err)
	}

	return orderID, tx.Commit(ctx)
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

// GetProcessingOrders returns all orders currently being fulfilled
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

// GetClientTelegramID returns the telegram_id for the client who owns an order
func (s *Store) GetClientTelegramID(ctx context.Context, orderID int64) (int64, error) {
	var tgID int64
	err := s.pool.QueryRow(ctx, `
		SELECT c.telegram_id FROM clients c
		JOIN orders o ON o.client_id = c.id
		WHERE o.id = $1
	`, orderID).Scan(&tgID)
	return tgID, err
}

// GetRefillableOrders returns completed Follower Booster orders not yet refilled
func (s *Store) GetRefillableOrders(ctx context.Context) ([]*models.Order, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT o.id, o.client_id, o.package_id, o.profile_link, o.total_kes,
		       o.status, o.wiz_order_ids, o.created_at, o.updated_at
		FROM orders o
		WHERE o.package_id = 'follower_booster'
		  AND o.status = 'completed'
		  AND o.created_at <= NOW() - INTERVAL '30 days'
		  AND NOT EXISTS (
		    SELECT 1 FROM refill_records r WHERE r.order_id = o.id
		  )
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
