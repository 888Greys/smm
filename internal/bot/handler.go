package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// State keys stored in memory per chat — replace with Redis/DB for multi-instance
type sessionState struct {
	Step      string
	PackageID string
}

var sessions = map[int64]*sessionState{}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update.CallbackQuery)
		return
	}
	if update.Message == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID
	sess := b.getSession(chatID)

	switch {
	case msg.Text == "/start":
		b.sendWelcome(chatID)
		sess.Step = ""

	case msg.Text == "/menu" || msg.Text == "/packages":
		b.sendPackageMenu(chatID)
		sess.Step = ""

	case msg.Text == "/balance" && b.isAdmin(msg.From.ID):
		b.sendBalance(ctx, chatID)

	case sess.Step == "awaiting_link":
		b.handleLinkSubmission(ctx, chatID, msg.From.ID, msg.Text, sess)

	default:
		b.sendPackageMenu(chatID)
	}
}

func (b *Bot) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	sess := b.getSession(chatID)

	// Acknowledge the button tap immediately
	b.api.Send(tgbotapi.NewCallback(cb.ID, ""))

	switch {
	case strings.HasPrefix(cb.Data, "pkg:"):
		pkgID := strings.TrimPrefix(cb.Data, "pkg:")
		pkg, ok := GetPackage(pkgID)
		if !ok {
			b.sendText(chatID, "Package not found.")
			return
		}
		sess.PackageID = pkgID
		sess.Step = "awaiting_link"
		b.sendText(chatID, fmt.Sprintf(
			"*%s* — KES %d\n_%s_\n\nPaste your profile/post link:",
			pkg.Name, pkg.PriceKES, pkg.Description,
		))

	case strings.HasPrefix(cb.Data, "approve:"):
		if !b.isAdmin(cb.From.ID) {
			return
		}
		orderID := strings.TrimPrefix(cb.Data, "approve:")
		b.approveOrder(ctx, chatID, cb.From.ID, orderID, cb.Message.MessageID)

	case strings.HasPrefix(cb.Data, "reject:"):
		if !b.isAdmin(cb.From.ID) {
			return
		}
		orderID := strings.TrimPrefix(cb.Data, "reject:")
		b.rejectOrder(ctx, chatID, cb.From.ID, orderID, cb.Message.MessageID)
	}
}

func (b *Bot) handleLinkSubmission(ctx context.Context, chatID, userID int64, link string, sess *sessionState) {
	link = strings.TrimSpace(link)
	if !isValidLink(link) {
		b.sendText(chatID, "That doesn't look like a valid link. Please paste the full URL (e.g. https://instagram.com/yourprofile).")
		return
	}

	pkg, ok := GetPackage(sess.PackageID)
	if !ok {
		b.sendText(chatID, "Something went wrong. Please start over with /menu.")
		sess.Step = ""
		return
	}

	// Create pending order + transaction in DB
	orderID, err := b.store.CreatePendingOrder(ctx, userID, pkg.ID, link, pkg.PriceKES)
	if err != nil {
		log.Printf("createPendingOrder: %v", err)
		b.sendText(chatID, "Could not create your order. Try again shortly.")
		return
	}

	// Confirm to client
	b.sendText(chatID, fmt.Sprintf(
		"Order received!\n\n*Package:* %s\n*Link:* %s\n*Amount:* KES %d\n\nSend payment via M-Pesa and share the confirmation code here or with our team.",
		pkg.Name, link, pkg.PriceKES,
	))

	// Notify admins for approval
	b.notifyAdmins(ctx, orderID, userID, pkg, link)
	sess.Step = ""
}

func (b *Bot) sendWelcome(chatID int64) {
	text := "Welcome to *AaPom SMM*!\n\nBoost your social media presence in minutes.\n\nUse /menu to see available packages."
	b.sendText(chatID, text)
}

func (b *Bot) sendPackageMenu(chatID int64) {
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, pkg := range Catalog {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s — KES %d", pkg.Name, pkg.PriceKES),
				"pkg:"+pkg.ID,
			),
		))
	}
	msg := tgbotapi.NewMessage(chatID, "Choose a package:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

func (b *Bot) sendBalance(ctx context.Context, chatID int64) {
	bal, err := b.wiz.GetBalance()
	if err != nil {
		b.sendText(chatID, fmt.Sprintf("Balance check failed: %v", err))
		return
	}
	b.sendText(chatID, fmt.Sprintf("SMMWiz balance: *%s %s*", bal.Balance, bal.Currency))
}

func (b *Bot) notifyAdmins(ctx context.Context, orderID, clientTelegramID int64, pkg Package, link string) {
	text := fmt.Sprintf(
		"New order #%d\nClient: `%d`\nPackage: *%s* (KES %d)\nLink: %s\n\nApprove once M-Pesa payment is confirmed.",
		orderID, clientTelegramID, pkg.Name, pkg.PriceKES, link,
	)
	approveBtn := tgbotapi.NewInlineKeyboardButtonData("Approve", fmt.Sprintf("approve:%d", orderID))
	rejectBtn := tgbotapi.NewInlineKeyboardButtonData("Reject", fmt.Sprintf("reject:%d", orderID))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(approveBtn, rejectBtn))

	for _, adminID := range b.adminIDs {
		msg := tgbotapi.NewMessage(adminID, text)
		msg.ReplyMarkup = keyboard
		msg.ParseMode = "Markdown"
		b.api.Send(msg)
	}
}

func (b *Bot) approveOrder(ctx context.Context, chatID, approverID int64, orderIDStr string, msgID int) {
	var orderID int64
	fmt.Sscanf(orderIDStr, "%d", &orderID)

	if err := b.store.ConfirmTransaction(ctx, orderID, approverID); err != nil {
		log.Printf("confirmTransaction: %v", err)
		b.sendText(chatID, "Failed to confirm transaction.")
		return
	}

	// Place all SMMWiz sub-orders asynchronously
	go b.fulfillOrder(context.Background(), orderID)

	// Edit the admin message to show it's done
	edit := tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("Order #%d approved. Fulfillment started.", orderID))
	b.api.Send(edit)
}

func (b *Bot) rejectOrder(ctx context.Context, chatID, approverID int64, orderIDStr string, msgID int) {
	var orderID int64
	fmt.Sscanf(orderIDStr, "%d", &orderID)

	if err := b.store.CancelOrder(ctx, orderID); err != nil {
		log.Printf("cancelOrder: %v", err)
		return
	}

	edit := tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("Order #%d rejected.", orderID))
	b.api.Send(edit)
}

func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

func (b *Bot) getSession(chatID int64) *sessionState {
	if sessions[chatID] == nil {
		sessions[chatID] = &sessionState{}
	}
	return sessions[chatID]
}

func (b *Bot) isAdmin(userID int64) bool {
	for _, id := range b.adminIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func isValidLink(link string) bool {
	return strings.HasPrefix(link, "https://") || strings.HasPrefix(link, "http://")
}
