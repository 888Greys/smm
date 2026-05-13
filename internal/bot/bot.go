package bot

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/aapom/smm/internal/megapay"
	"github.com/aapom/smm/internal/models"
	"github.com/aapom/smm/internal/smmpanel"
)

type Store interface {
	UpsertClient(ctx context.Context, telegramID int64) (bool, error)
	CreatePendingOrder(ctx context.Context, clientTelegramID int64, packageID, link string, amountKES int, referralCode string) (int64, error)
	SaveSTKRequest(ctx context.Context, orderID int64, phone, stkRequestID string) error
	ConfirmTransaction(ctx context.Context, orderID, confirmedBy int64) error
	CancelOrder(ctx context.Context, orderID int64) error
	GetOrder(ctx context.Context, orderID int64) (*models.Order, error)
	GetClientTelegramID(ctx context.Context, orderID int64) (int64, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status models.OrderStatus, wizIDs []int64) error
	SaveRefill(ctx context.Context, orderID, wizOrderID, wizRefillID int64) error
	GetOrCreateReferralCode(ctx context.Context, telegramID int64) (string, error)
	GetCreditBalance(ctx context.Context, telegramID int64) (int, error)
	GetStats(ctx context.Context) (*models.DailyStats, error)
}

type Package = models.Package

type Bot struct {
	api            *tgbotapi.BotAPI
	wiz            *smmpanel.Client
	pay            *megapay.Client
	store          Store
	adminIDs       []int64
	proofChannelID int64
	notifier       *AdminNotifier // nil = disabled
}

func New(token string, wiz *smmpanel.Client, pay *megapay.Client, store Store, adminIDs []int64, proofChannelID int64, notifier *AdminNotifier) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api, wiz: wiz, pay: pay, store: store, adminIDs: adminIDs, proofChannelID: proofChannelID, notifier: notifier}, nil
}

func (b *Bot) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	log.Printf("bot started: @%s", b.api.Self.UserName)

	// Admin bot handles ALL admin-facing callbacks (fulfill:/reject:)
	if b.notifier != nil {
		go b.notifier.StartCallbackListener(ctx, func(cb *tgbotapi.CallbackQuery) {
			b.notifier.AckCallback(cb.ID)
			go b.handleAdminCallback(ctx, cb)
		})
	}

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

// followerServices are SMMWiz service IDs that deliver followers (auto drip-feed eligible).
var followerServices = map[int]bool{5760: true, 5440: true}

// applyAutoDrip adds drip-feed parameters to follower components with >1000 quantity
// that don't already have explicit runs set. Rate: ~300-500 followers/day.
func applyAutoDrip(comp models.PackageComponent) models.PackageComponent {
	if followerServices[comp.ServiceID] && comp.Quantity > 1000 && comp.Runs == 0 {
		runs := comp.Quantity / 400 // ~400/day
		if runs < 2 {
			runs = 2
		}
		if runs > 30 {
			runs = 30
		}
		comp.Runs = runs
		comp.Interval = 1440 // 24 hours between runs
	}
	return comp
}

// FulfillOrder places all SMMWiz sub-orders for orderID.
// sendText is used to send progress messages to the client (may be nil for web orders).
// notifier may be nil if no admin bot is configured.
func FulfillOrder(ctx context.Context, store Store, wiz *smmpanel.Client, sendText func(int64, string), notifier *AdminNotifier, orderID int64) {
	order, err := store.GetOrder(ctx, orderID)
	if err != nil {
		log.Printf("FulfillOrder getOrder %d: %v", orderID, err)
		return
	}

	pkg, ok := GetPackage(order.PackageID)
	if !ok {
		log.Printf("FulfillOrder unknown package %s", order.PackageID)
		return
	}

	clientTgID, _ := store.GetClientTelegramID(ctx, orderID)
	total := len(pkg.Components)
	multiStep := total > 1 && clientTgID > 0 && sendText != nil

	var wizIDs []int64
	for i, comp := range pkg.Components {
		comp = applyAutoDrip(comp)

		if multiStep {
			sendText(clientTgID, fmt.Sprintf(
				"⚡ *Placing component %d/%d…*\n_%s_",
				i+1, total, componentLabel(comp),
			))
		}

		req := smmpanel.OrderRequest{
			Service:  comp.ServiceID,
			Link:     order.ProfileLink,
			Quantity: comp.Quantity,
		}
		if comp.Runs > 0 {
			req.Runs = comp.Runs
			req.Interval = comp.Interval
		}

		resp, err := wiz.AddOrder(req)
		if err != nil {
			log.Printf("FulfillOrder AddOrder (order %d service %d): %v", orderID, comp.ServiceID, err)
			store.UpdateOrderStatus(ctx, orderID, models.StatusFailed, wizIDs)
			if clientTgID > 0 && sendText != nil {
				sendText(clientTgID, "⚠️ Your order could not be placed. Please contact @workratew for support.")
			}
			return
		}
		wizIDs = append(wizIDs, resp.Order)
		log.Printf("order %d → wiz order %d placed (service %d qty %d)", orderID, resp.Order, comp.ServiceID, comp.Quantity)
	}

	if err := store.UpdateOrderStatus(ctx, orderID, models.StatusProcessing, wizIDs); err != nil {
		log.Printf("FulfillOrder updateStatus %d: %v", orderID, err)
	}

	if notifier != nil {
		notifier.NotifyOrderFulfilled(orderID, pkg, wizIDs)
	}
}

func (b *Bot) fulfillOrder(ctx context.Context, orderID int64) {
	FulfillOrder(ctx, b.store, b.wiz, b.sendRaw, b.notifier, orderID)
}

// componentLabel returns a human-readable label for a package component.
func componentLabel(comp models.PackageComponent) string {
	labels := map[int]string{
		5760: "TikTok Followers",
		9121: "TikTok Views",
		2699: "TikTok Likes",
		5440: "Instagram Followers",
		2916: "Instagram Likes",
		7494: "YouTube Subscribers",
		6003: "YouTube Views",
	}
	if l, ok := labels[comp.ServiceID]; ok {
		return fmt.Sprintf("%s × %d", l, comp.Quantity)
	}
	return fmt.Sprintf("Service %d × %d", comp.ServiceID, comp.Quantity)
}

func (b *Bot) sendRaw(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	b.api.Send(m)
}
