package bot

import (
	"context"
	"fmt"
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
	wiz            *smmwiz.Client
	pay            *megapay.Client
	store          Store
	adminIDs       []int64
	proofChannelID int64
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

// followerServices are SMMWiz service IDs that deliver followers (auto drip-feed eligible).
var followerServices = map[int]bool{18612: true, 20888: true}

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

	// Look up client Telegram ID for progress notifications (0 = web order, skip)
	clientTgID, _ := b.store.GetClientTelegramID(ctx, orderID)
	total := len(pkg.Components)
	multiStep := total > 1 && clientTgID > 0

	var wizIDs []int64
	for i, comp := range pkg.Components {
		comp = applyAutoDrip(comp)

		if multiStep {
			b.sendRaw(clientTgID, fmt.Sprintf(
				"⚡ *Placing component %d/%d…*\n_%s_",
				i+1, total, componentLabel(comp),
			))
		}

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
			if clientTgID > 0 {
				b.sendRaw(clientTgID, "⚠️ Your order could not be placed. Please contact @workratew for support.")
			}
			return
		}
		wizIDs = append(wizIDs, resp.Order)
		log.Printf("order %d → wiz order %d placed (service %d qty %d)", orderID, resp.Order, comp.ServiceID, comp.Quantity)
	}

	if err := b.store.UpdateOrderStatus(ctx, orderID, models.StatusProcessing, wizIDs); err != nil {
		log.Printf("fulfillOrder updateStatus %d: %v", orderID, err)
	}
}

// componentLabel returns a human-readable label for a package component.
func componentLabel(comp models.PackageComponent) string {
	labels := map[int]string{
		18612: "TikTok Followers",
		19967: "TikTok Views",
		12350: "TikTok Likes",
		20888: "Instagram Followers",
		19909: "Instagram Likes",
		9727:  "YouTube Subscribers",
		19646: "YouTube Views",
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
