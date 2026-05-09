package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminNotifier sends all admin-facing events through a dedicated admin bot.
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

// send sends a plain Markdown message to the admin chat.
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

// SendWithButtons sends a message with an inline keyboard to the admin chat.
func (n *AdminNotifier) SendWithButtons(text string, kb tgbotapi.InlineKeyboardMarkup) {
	if n == nil {
		return
	}
	msg := tgbotapi.NewMessage(n.chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	if _, err := n.api.Send(msg); err != nil {
		log.Printf("adminNotifier sendWithButtons: %v", err)
	}
}

// EditMessage edits an existing message in the admin chat.
func (n *AdminNotifier) EditMessage(msgID int, text string) {
	if n == nil {
		return
	}
	edit := tgbotapi.NewEditMessageText(n.chatID, msgID, text)
	edit.ParseMode = "Markdown"
	if _, err := n.api.Send(edit); err != nil {
		log.Printf("adminNotifier editMessage: %v", err)
	}
}

// AckCallback silently acknowledges a callback query.
func (n *AdminNotifier) AckCallback(callbackID string) {
	if n == nil {
		return
	}
	n.api.Send(tgbotapi.NewCallback(callbackID, ""))
}

// StartCallbackListener runs the admin bot update loop and calls onCallback for
// every inline button tap. Run as a goroutine alongside the main bot.
func (n *AdminNotifier) StartCallbackListener(ctx context.Context, onCallback func(*tgbotapi.CallbackQuery)) {
	if n == nil {
		return
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := n.api.GetUpdatesChan(u)
	log.Printf("admin bot callback listener started")
	for {
		select {
		case <-ctx.Done():
			n.api.StopReceivingUpdates()
			return
		case update := <-updates:
			if update.CallbackQuery != nil {
				onCallback(update.CallbackQuery)
			}
		}
	}
}

// ── Traffic notifications ─────────────────────────────────────────────────────

func (n *AdminNotifier) NotifyNewUser(userID int64, username, firstName string, referralCode string) {
	name := firstName
	if username != "" {
		name = "@" + username
	}
	ref := ""
	if referralCode != "" {
		ref = fmt.Sprintf("\n🔗 Referred by: `%s`", referralCode)
	}
	n.send(fmt.Sprintf(
		"👤 *New User*\n\nName: %s\nID: `%d`\nTime: %s%s",
		name, userID, time.Now().Format("02 Jan 15:04 MST"), ref,
	))
}

func (n *AdminNotifier) NotifyReturningUser(userID int64, username, firstName string) {
	name := firstName
	if username != "" {
		name = "@" + username
	}
	n.send(fmt.Sprintf(
		"🔄 *Returning User*\n\nName: %s\nID: `%d`\nTime: %s",
		name, userID, time.Now().Format("02 Jan 15:04 MST"),
	))
}

func (n *AdminNotifier) NotifyPackageSelected(userID int64, username string, pkg Package) {
	name := fmt.Sprintf("`%d`", userID)
	if username != "" {
		name = "@" + username
	}
	n.send(fmt.Sprintf(
		"%s *Package Viewed*\n\nUser: %s\nPackage: *%s*\nPrice: KES %d\nTime: %s",
		platformEmoji(string(pkg.Platform)),
		name, pkg.Name, pkg.PriceKES,
		time.Now().Format("02 Jan 15:04 MST"),
	))
}

func (n *AdminNotifier) NotifyOrderCreated(orderID, userID int64, username string, pkg Package, phone string) {
	name := fmt.Sprintf("`%d`", userID)
	if username != "" {
		name = "@" + username
	}
	n.send(fmt.Sprintf(
		"💳 *Order Created — Awaiting Payment*\n\n"+
			"Order: #%d\nUser: %s\nPackage: *%s*\nAmount: KES %d\nPhone: `%s`\nTime: %s",
		orderID, name, pkg.Name, pkg.PriceKES, phone,
		time.Now().Format("02 Jan 15:04 MST"),
	))
}

// NotifyPaymentConfirmedWithFulfill sends the payment confirmation with
// Fulfill / Reject inline buttons. The admin bot receives the callback tap.
func (n *AdminNotifier) NotifyPaymentConfirmedWithFulfill(orderID int64, pkg Package, clientTgID int64, profileLink, mpesaRef string) {
	if n == nil {
		return
	}
	display := profileLink
	if len(display) > 50 {
		display = display[:47] + "..."
	}
	text := fmt.Sprintf(
		"💰 *Payment Confirmed — Order #%d*\n\n"+
			"📦 %s\n"+
			"💰 KES %d\n"+
			"🔗 %s\n"+
			"👤 Client: `%d`\n"+
			"📱 M-Pesa ref: `%s`\n"+
			"🕐 %s\n\n"+
			"Tap *Fulfill* to send to SMMWiz and start delivery.",
		orderID, pkg.Name, pkg.PriceKES,
		display, clientTgID, mpesaRef,
		time.Now().Format("02 Jan 15:04 MST"),
	)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Fulfill Order", fmt.Sprintf("fulfill:%d", orderID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Reject", fmt.Sprintf("reject:%d", orderID)),
		),
	)
	n.SendWithButtons(text, kb)
}

func (n *AdminNotifier) NotifyOrderFulfilled(orderID int64, pkg Package, wizIDs []int64) {
	n.send(fmt.Sprintf(
		"✅ *Order Fulfilled*\n\nOrder: #%d\nPackage: *%s*\nWiz IDs: %v\nTime: %s",
		orderID, pkg.Name, wizIDs,
		time.Now().Format("02 Jan 15:04 MST"),
	))
}
