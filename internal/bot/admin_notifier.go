package bot

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminNotifier sends traffic and order events to a separate admin bot/chat.
type AdminNotifier struct {
	api    *tgbotapi.BotAPI
	chatID int64
}

func NewAdminNotifier(token string, chatID int64) (*AdminNotifier, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("admin bot init: %w", err)
	}
	log.Printf("admin notifier ready: @%s → chat %d", api.Self.UserName, chatID)
	return &AdminNotifier{api: api, chatID: chatID}, nil
}

func (n *AdminNotifier) send(text string) {
	if n == nil {
		return
	}
	msg := tgbotapi.NewMessage(n.chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := n.api.Send(msg); err != nil {
		log.Printf("adminNotifier send: %v", err)
	}
}

// NotifyNewUser fires when a user sends /start for the first time.
func (n *AdminNotifier) NotifyNewUser(userID int64, username, firstName string, referralCode string) {
	name := firstName
	if username != "" {
		name = "@" + username
	}
	ref := ""
	if referralCode != "" {
		ref = fmt.Sprintf("\n🔗 Referred by code: `%s`", referralCode)
	}
	n.send(fmt.Sprintf(
		"👤 *New User*\n\n"+
			"Name: %s\n"+
			"ID: `%d`\n"+
			"Time: %s%s",
		name, userID, time.Now().Format("02 Jan 15:04 MST"), ref,
	))
}

// NotifyReturningUser fires when an existing user sends /start again.
func (n *AdminNotifier) NotifyReturningUser(userID int64, username, firstName string) {
	name := firstName
	if username != "" {
		name = "@" + username
	}
	n.send(fmt.Sprintf(
		"🔄 *Returning User*\n\n"+
			"Name: %s\n"+
			"ID: `%d`\n"+
			"Time: %s",
		name, userID, time.Now().Format("02 Jan 15:04 MST"),
	))
}

// NotifyPackageSelected fires when a user taps a package in the shop.
func (n *AdminNotifier) NotifyPackageSelected(userID int64, username string, pkg Package) {
	name := fmt.Sprintf("`%d`", userID)
	if username != "" {
		name = "@" + username
	}
	n.send(fmt.Sprintf(
		"%s *Package Viewed*\n\n"+
			"User: %s\n"+
			"Package: *%s*\n"+
			"Price: KES %d\n"+
			"Time: %s",
		platformEmoji(string(pkg.Platform)),
		name, pkg.Name, pkg.PriceKES,
		time.Now().Format("02 Jan 15:04 MST"),
	))
}

// NotifyOrderCreated fires when an order is created and STK push sent.
func (n *AdminNotifier) NotifyOrderCreated(orderID int64, userID int64, username string, pkg Package, phone string) {
	name := fmt.Sprintf("`%d`", userID)
	if username != "" {
		name = "@" + username
	}
	n.send(fmt.Sprintf(
		"💳 *Order Created — Awaiting Payment*\n\n"+
			"Order: #%d\n"+
			"User: %s\n"+
			"Package: *%s*\n"+
			"Amount: KES %d\n"+
			"Phone: `%s`\n"+
			"Time: %s",
		orderID, name, pkg.Name, pkg.PriceKES, phone,
		time.Now().Format("02 Jan 15:04 MST"),
	))
}

// NotifyPaymentConfirmed fires when M-Pesa payment is confirmed (called from worker).
func (n *AdminNotifier) NotifyPaymentConfirmed(orderID int64, amountKES int, receipt string) {
	n.send(fmt.Sprintf(
		"💰 *Payment Confirmed*\n\n"+
			"Order: #%d\n"+
			"Amount: KES %d\n"+
			"Receipt: `%s`\n"+
			"Time: %s",
		orderID, amountKES, receipt,
		time.Now().Format("02 Jan 15:04 MST"),
	))
}

// NotifyOrderFulfilled fires when SMMWiz order is placed successfully.
func (n *AdminNotifier) NotifyOrderFulfilled(orderID int64, pkg Package, wizIDs []int64) {
	n.send(fmt.Sprintf(
		"✅ *Order Fulfilled*\n\n"+
			"Order: #%d\n"+
			"Package: *%s*\n"+
			"Wiz IDs: %v\n"+
			"Time: %s",
		orderID, pkg.Name, wizIDs,
		time.Now().Format("02 Jan 15:04 MST"),
	))
}
