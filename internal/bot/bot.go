package bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/aapom/smm/internal/megapay"
	"github.com/aapom/smm/internal/models"
	"github.com/aapom/smm/internal/smmwiz"
)

type Store interface {
	UpsertClient(ctx context.Context, telegramID int64) error
	CreatePendingOrder(ctx context.Context, clientTelegramID int64, packageID, link string, amountKES int, referralCode string) (int64, error)
	SaveSTKRequest(ctx context.Context, orderID int64, phone, stkRequestID string) error
	ConfirmTransaction(ctx context.Context, orderID, confirmedBy int64) error
	CancelOrder(ctx context.Context, orderID int64) error
	GetOrder(ctx context.Context, orderID int64) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status models.OrderStatus, wizIDs []int64) error
	SaveRefill(ctx context.Context, orderID, wizOrderID, wizRefillID int64) error
	GetOrCreateReferralCode(ctx context.Context, telegramID int64) (string, error)
	GetCreditBalance(ctx context.Context, telegramID int64) (int, error)
	GetStats(ctx context.Context) (*models.DailyStats, error)
}

type Package = models.Package

type Bot struct {
	api            *tgbotapi.BotAPI
	wiz            *smmwiz.Client
	pay            *megapay.Client
	store          Store
	adminIDs       []int64
	proofChannelID int64 // optional: public channel for social proof posts
}

func New(token string, wiz *smmwiz.Client, pay *megapay.Client, store Store, adminIDs []int64, proofChannelID int64) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api, wiz: wiz, pay: pay, store: store, adminIDs: adminIDs, proofChannelID: proofChannelID}, nil
}

func (b *Bot) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	log.Printf("bot started: @%s", b.api.Self.UserName)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update := <-updates:
			go b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) fulfillOrder(ctx context.Context, orderID int64) {
	order, err := b.store.GetOrder(ctx, orderID)
	if err != nil {
		log.Printf("fulfillOrder getOrder %d: %v", orderID, err)
		return
	}

	pkg, ok := GetPackage(order.PackageID)
	if !ok {
		log.Printf("fulfillOrder unknown package %s", order.PackageID)
		return
	}

	var wizIDs []int64
	for _, comp := range pkg.Components {
		req := smmwiz.OrderRequest{
			Service:  comp.ServiceID,
			Link:     order.ProfileLink,
			Quantity: comp.Quantity,
		}
		if comp.Runs > 0 {
			req.Runs = comp.Runs
			req.Interval = comp.Interval
		}

		resp, err := b.wiz.AddOrder(req)
		if err != nil {
			log.Printf("fulfillOrder AddOrder (order %d service %d): %v", orderID, comp.ServiceID, err)
			b.store.UpdateOrderStatus(ctx, orderID, models.StatusFailed, wizIDs)
			return
		}
		wizIDs = append(wizIDs, resp.Order)
		log.Printf("order %d → wiz order %d placed", orderID, resp.Order)
	}

	if err := b.store.UpdateOrderStatus(ctx, orderID, models.StatusProcessing, wizIDs); err != nil {
		log.Printf("fulfillOrder updateStatus %d: %v", orderID, err)
	}
}
