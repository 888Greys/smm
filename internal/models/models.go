package models

import "time"

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusCompleted  OrderStatus = "completed"
	StatusPartial    OrderStatus = "partial"
	StatusCancelled  OrderStatus = "cancelled"
	StatusFailed     OrderStatus = "failed"
)

type Platform string

const (
	PlatformTikTok   Platform = "tiktok"
	PlatformInstagram Platform = "instagram"
	PlatformYouTube  Platform = "youtube"
)

// Package defines a retail service bundle sold to clients
type Package struct {
	ID          string
	Name        string
	Platform    Platform
	PriceKES    int
	MarginKES   int  // retail price minus wholesale cost (at ~130 KES/USD)
	Refillable  bool // whether this package gets a 30-day refill
	Description string
	Components  []PackageComponent
}

// DailyStats is returned by Store.GetStats for the admin dashboard
type DailyStats struct {
	Lines            []PackageStatLine
	PendingOrders    int
	ProcessingOrders int
	CompletedOrders  int
	TotalOrders      int
}

type PackageStatLine struct {
	PackageID  string
	OrderCount int
	RevenueKES int
}

// PackageComponent maps to a single SMMWiz service call within a package
type PackageComponent struct {
	ServiceID int
	Quantity  int
	Runs      int // for drip-feed: number of runs
	Interval  int // for drip-feed: minutes between runs
}

// Order is a client order stored in our DB
type Order struct {
	ID            int64
	ClientID      int64
	PackageID     string
	ProfileLink   string
	TotalKES      int
	Status        OrderStatus
	WizOrderIDs   []int64 // one per PackageComponent
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Client is a Telegram user who has placed at least one order
type Client struct {
	ID         int64
	TelegramID int64
	Username   string
	Phone      string
	CreatedAt  time.Time
}

// Transaction records a payment event
type Transaction struct {
	ID            int64
	OrderID       int64
	AmountKES     int
	MpesaRef      string
	Confirmed     bool
	ConfirmedBy   int64 // Telegram user ID of approver
	ConfirmedAt   *time.Time
	CreatedAt     time.Time
}

// RefillRecord tracks a refill request for an order
type RefillRecord struct {
	ID          int64
	OrderID     int64
	WizOrderID  int64
	WizRefillID int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
