package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/aapom/smm/internal/db"
	"github.com/aapom/smm/internal/models"
	"github.com/aapom/smm/internal/smmwiz"
)

func main() {
	godotenv.Load()

	store, err := db.NewStore(context.Background(), mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer store.Close()

	wiz := smmwiz.New(mustEnv("SMMWIZ_API_KEY"))

	tgAPI, err := tgbotapi.NewBotAPI(mustEnv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("telegram: %v", err)
	}

	balanceThreshold, _ := strconv.ParseFloat(os.Getenv("BALANCE_ALERT_THRESHOLD"), 64)
	if balanceThreshold == 0 {
		balanceThreshold = 5.0
	}

	adminIDs := parseAdminIDs(mustEnv("ADMIN_TELEGRAM_IDS"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pollTicker := time.NewTicker(30 * time.Minute)
	refillTicker := time.NewTicker(24 * time.Hour)
	balanceTicker := time.NewTicker(12 * time.Hour)

	log.Println("worker started")

	// Run immediately on start, then on ticker
	pollOrders(ctx, store, wiz, tgAPI)
	checkBalance(ctx, wiz, tgAPI, adminIDs, balanceThreshold)

	for {
		select {
		case <-ctx.Done():
			log.Println("worker stopped")
			return
		case <-pollTicker.C:
			pollOrders(ctx, store, wiz, tgAPI)
		case <-refillTicker.C:
			triggerRefills(ctx, store, wiz)
		case <-balanceTicker.C:
			checkBalance(ctx, wiz, tgAPI, adminIDs, balanceThreshold)
		}
	}
}

func pollOrders(ctx context.Context, store *db.Store, wiz *smmwiz.Client, tg *tgbotapi.BotAPI) {
	orders, err := store.GetProcessingOrders(ctx)
	if err != nil {
		log.Printf("pollOrders fetch: %v", err)
		return
	}
	if len(orders) == 0 {
		return
	}

	// Collect all wiz order IDs across all orders
	wizToOrder := map[int64]int64{} // wizOrderID → our orderID
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

	// Track which orders had all components complete
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
			log.Printf("order %d wiz_order %d status: %s (remains: %s)", orderID, wizID, s.Status, s.Remains)
		}
	}

	for orderID := range orderComplete {
		if orderFailed[orderID] {
			continue // partial — leave in processing, keep polling
		}
		if err := store.UpdateOrderStatus(ctx, orderID, models.StatusCompleted, nil); err != nil {
			log.Printf("updateStatus completed %d: %v", orderID, err)
			continue
		}
		notifyClient(ctx, store, tg, orderID, "Your order is complete! Check your profile.")
		log.Printf("order %d completed", orderID)
	}

	for orderID := range orderFailed {
		if err := store.UpdateOrderStatus(ctx, orderID, models.StatusPartial, nil); err != nil {
			log.Printf("updateStatus partial %d: %v", orderID, err)
		}
		log.Printf("order %d marked partial", orderID)
	}
}

func triggerRefills(ctx context.Context, store *db.Store, wiz *smmwiz.Client) {
	orders, err := store.GetRefillableOrders(ctx)
	if err != nil {
		log.Printf("triggerRefills fetch: %v", err)
		return
	}

	for _, o := range orders {
		for _, wizID := range o.WizOrderIDs {
			resp, err := wiz.Refill(wizID)
			if err != nil {
				log.Printf("refill order %d wiz %d: %v", o.ID, wizID, err)
				continue
			}
			if err := store.SaveRefill(ctx, o.ID, wizID, resp.Refill); err != nil {
				log.Printf("saveRefill order %d: %v", o.ID, err)
			}
			log.Printf("refill triggered: order %d wiz_order %d → refill %d", o.ID, wizID, resp.Refill)
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
		msg := fmt.Sprintf("LOW BALANCE ALERT\nSMMWiz wallet: *%s %s*\nTop up now or orders will fail.", bal.Balance, bal.Currency)
		for _, id := range adminIDs {
			m := tgbotapi.NewMessage(id, msg)
			m.ParseMode = "Markdown"
			tg.Send(m)
		}
	}
}

func notifyClient(ctx context.Context, store *db.Store, tg *tgbotapi.BotAPI, orderID int64, text string) {
	tgID, err := store.GetClientTelegramID(ctx, orderID)
	if err != nil {
		log.Printf("notifyClient getID %d: %v", orderID, err)
		return
	}
	m := tgbotapi.NewMessage(tgID, text)
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
