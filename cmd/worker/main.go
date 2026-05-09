package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/aapom/smm/internal/bot"
	"github.com/aapom/smm/internal/db"
	"github.com/aapom/smm/internal/megapay"
	"github.com/aapom/smm/internal/models"
	"github.com/aapom/smm/internal/smmwiz"
)

func main() {
	godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := db.NewStore(ctx, mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer store.Close()

	wiz := smmwiz.New(mustEnv("SMMWIZ_API_KEY"))
	pay := megapay.New(mustEnv("MEGAPAY_API_KEY"), mustEnv("MEGAPAY_EMAIL"))

	tgAPI, err := tgbotapi.NewBotAPI(mustEnv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("telegram: %v", err)
	}

	balanceThreshold, _ := strconv.ParseFloat(os.Getenv("BALANCE_ALERT_THRESHOLD"), 64)
	if balanceThreshold == 0 {
		balanceThreshold = 5.0
	}
	adminIDs := parseAdminIDs(mustEnv("ADMIN_TELEGRAM_IDS"))

	var proofChannelID int64
	if ch := os.Getenv("SOCIAL_PROOF_CHANNEL_ID"); ch != "" {
		proofChannelID, _ = strconv.ParseInt(ch, 10, 64)
	}

	paymentTicker := time.NewTicker(10 * time.Second)
	orderTicker := time.NewTicker(30 * time.Minute)
	refillTicker := time.NewTicker(24 * time.Hour)
	balanceTicker := time.NewTicker(12 * time.Hour)

	log.Println("worker started")
	checkBalance(ctx, wiz, tgAPI, adminIDs, balanceThreshold)
	pollOrders(ctx, store, wiz, tgAPI, proofChannelID)

	for {
		select {
		case <-ctx.Done():
			log.Println("worker stopped")
			return
		case <-paymentTicker.C:
			pollPayments(ctx, store, pay, wiz, tgAPI, adminIDs, proofChannelID)
		case <-orderTicker.C:
			pollOrders(ctx, store, wiz, tgAPI, proofChannelID)
		case <-refillTicker.C:
			triggerRefills(ctx, store, wiz)
		case <-balanceTicker.C:
			checkBalance(ctx, wiz, tgAPI, adminIDs, balanceThreshold)
		}
	}
}

func pollPayments(ctx context.Context, store *db.Store, pay *megapay.Client, wiz *smmwiz.Client, tg *tgbotapi.BotAPI, adminIDs []int64, proofChannelID int64) {
	txns, err := store.GetPendingSTKTransactions(ctx)
	if err != nil {
		log.Printf("getPendingSTK: %v", err)
		return
	}

	for _, txn := range txns {
		status, err := pay.CheckStatus(txn.STKRequestID)
		if err != nil {
			log.Printf("checkStatus txn %s: %v", txn.STKRequestID, err)
			continue
		}

		log.Printf("payment poll: order %d txnStatus=%s txnCode=%s receipt=%s",
			txn.OrderID, status.TransactionStatus, status.TransactionCode, status.TransactionReceipt)

		switch status.TransactionStatus {
		case "Cancelled":
			store.CancelOrder(ctx, txn.OrderID)
			notifyClient(ctx, store, tg, txn.OrderID, "❌ Payment was cancelled. Send /start to try again.")
			log.Printf("order %d payment cancelled by user", txn.OrderID)
			continue
		case "Completed":
			if status.TransactionCode != "0" {
				continue
			}
		default:
			continue
		}

		if err := store.ConfirmTransaction(ctx, txn.OrderID, 0); err != nil {
			log.Printf("confirmTransaction order %d: %v", txn.OrderID, err)
			continue
		}

		notifyAdmins(tg, adminIDs, fmt.Sprintf("💰 Payment confirmed for Order #%d — KES %d (receipt: %s)",
			txn.OrderID, txn.AmountKES, status.TransactionReceipt))

		go fulfillOrder(ctx, store, wiz, tg, txn.OrderID, proofChannelID)
	}
}

func fulfillOrder(ctx context.Context, store *db.Store, wiz *smmwiz.Client, tg *tgbotapi.BotAPI, orderID int64, proofChannelID int64) {
	order, err := store.GetOrder(ctx, orderID)
	if err != nil {
		log.Printf("fulfillOrder getOrder %d: %v", orderID, err)
		return
	}

	pkg, ok := bot.GetPackage(order.PackageID)
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

		resp, err := wiz.AddOrder(req)
		if err != nil {
			log.Printf("fulfillOrder AddOrder (order %d service %d): %v", orderID, comp.ServiceID, err)
			store.UpdateOrderStatus(ctx, orderID, models.StatusFailed, wizIDs)
			notifyClient(ctx, store, tg, orderID, "⚠️ Your order could not be placed. Please contact support.")
			return
		}
		wizIDs = append(wizIDs, resp.Order)
		log.Printf("order %d → wiz order %d placed", orderID, resp.Order)
	}

	store.UpdateOrderStatus(ctx, orderID, models.StatusProcessing, wizIDs)

	notifyClient(ctx, store, tg, orderID, fmt.Sprintf(
		"✅ *Payment confirmed!*\n\nYour order #%d has been placed and delivery has started.\n\nYou'll be notified when it's complete.", orderID,
	))
}

func pollOrders(ctx context.Context, store *db.Store, wiz *smmwiz.Client, tg *tgbotapi.BotAPI, proofChannelID int64) {
	orders, err := store.GetProcessingOrders(ctx)
	if err != nil {
		log.Printf("pollOrders fetch: %v", err)
		return
	}
	if len(orders) == 0 {
		return
	}

	wizToOrder := map[int64]int64{}
	var allWizIDs []int64
	for _, o := range orders {
		for _, wid := range o.WizOrderIDs {
			wizToOrder[wid] = o.ID
			allWizIDs = append(allWizIDs, wid)
		}
	}

	statuses, err := wiz.GetMultiStatus(allWizIDs)
	if err != nil {
		log.Printf("pollOrders multiStatus: %v", err)
		return
	}

	orderComplete := map[int64]bool{}
	orderFailed := map[int64]bool{}

	for wizIDStr, s := range statuses {
		var wizID int64
		fmt.Sscanf(wizIDStr, "%d", &wizID)
		orderID := wizToOrder[wizID]

		switch s.Status {
		case "Completed":
			orderComplete[orderID] = true
		case "Partial", "Canceled":
			orderFailed[orderID] = true
		}
	}

	for orderID := range orderComplete {
		if orderFailed[orderID] {
			continue
		}
		store.UpdateOrderStatus(ctx, orderID, models.StatusCompleted, nil)
		log.Printf("order %d completed", orderID)

		// Award referral credit if applicable
		if referrerTgID, err := store.AwardReferralCredit(ctx, orderID); err != nil {
			log.Printf("awardReferralCredit order %d: %v", orderID, err)
		} else if referrerTgID > 0 {
			send(tg, referrerTgID, "🎉 *Referral bonus!*\n\nYour friend just completed their first order.\nKES 50 credit has been added to your account!")
			log.Printf("referral credit awarded for order %d → tg %d", orderID, referrerTgID)
		}

		// Build completion message with optional upsell
		completionMsg := buildCompletionMessage(orderID, store, ctx)
		notifyClient(ctx, store, tg, orderID, completionMsg)

		// Social proof channel post
		if proofChannelID != 0 {
			order, err := store.GetOrder(ctx, orderID)
			if err == nil {
				if pkg, ok := bot.GetPackage(order.PackageID); ok {
					postSocialProof(tg, proofChannelID, pkg)
				}
			}
		}
	}
}

func buildCompletionMessage(orderID int64, store *db.Store, ctx context.Context) string {
	base := "🎉 *Order complete!*\n\nYour followers are live — check your profile!"

	order, err := store.GetOrder(ctx, orderID)
	if err != nil {
		return base
	}

	upsell, ok := bot.UpsellTarget(order.PackageID)
	if !ok {
		return base
	}

	return base + fmt.Sprintf(
		"\n\n💥 *Ready for more?*\nUpgrade to *%s* for just *KES %d* and 10× your growth!\n\nTap 🛍 *Shop* to order now.",
		upsell.Name, upsell.PriceKES,
	)
}

func postSocialProof(tg *tgbotapi.BotAPI, channelID int64, pkg models.Package) {
	emoji := "📱"
	switch pkg.Platform {
	case models.PlatformTikTok:
		emoji = "🎵"
	case models.PlatformInstagram:
		emoji = "📸"
	case models.PlatformYouTube:
		emoji = "▶️"
	}
	text := fmt.Sprintf(
		"✅ *Another order delivered!*\n\n%s *%s*\n📦 %s\n\n⚡️ Delivered fast. No drops.\n\n🛍 Get yours → @AaPomSMM",
		emoji, pkg.Name, pkg.Description,
	)
	send(tg, channelID, text)
}

func triggerRefills(ctx context.Context, store *db.Store, wiz *smmwiz.Client) {
	orders, err := store.GetRefillableOrders(ctx, bot.RefillablePackageIDs())
	if err != nil {
		log.Printf("triggerRefills: %v", err)
		return
	}
	for _, o := range orders {
		for _, wizID := range o.WizOrderIDs {
			resp, err := wiz.Refill(wizID)
			if err != nil {
				log.Printf("refill order %d wiz %d: %v", o.ID, wizID, err)
				continue
			}
			store.SaveRefill(ctx, o.ID, wizID, resp.Refill)
			log.Printf("refill triggered: order %d → refill %d", o.ID, resp.Refill)
		}
	}
}

func checkBalance(ctx context.Context, wiz *smmwiz.Client, tg *tgbotapi.BotAPI, adminIDs []int64, threshold float64) {
	bal, err := wiz.GetBalance()
	if err != nil {
		log.Printf("balance check: %v", err)
		return
	}
	balance, _ := strconv.ParseFloat(bal.Balance, 64)
	log.Printf("SMMWiz balance: %s %s", bal.Balance, bal.Currency)
	if balance < threshold {
		notifyAdmins(tg, adminIDs, fmt.Sprintf("⚠️ *LOW BALANCE*\nSMMWiz wallet: *%s %s*\nTop up now or orders will fail.", bal.Balance, bal.Currency))
	}
}

func notifyClient(ctx context.Context, store *db.Store, tg *tgbotapi.BotAPI, orderID int64, text string) {
	tgID, err := store.GetClientTelegramID(ctx, orderID)
	if err != nil {
		log.Printf("notifyClient getID %d: %v", orderID, err)
		return
	}
	send(tg, tgID, text)
}

func notifyAdmins(tg *tgbotapi.BotAPI, adminIDs []int64, text string) {
	for _, id := range adminIDs {
		send(tg, id, text)
	}
}

func send(tg *tgbotapi.BotAPI, chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	tg.Send(m)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func parseAdminIDs(s string) []int64 {
	var ids []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if id, err := strconv.ParseInt(part, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
