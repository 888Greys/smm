package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type sessionState struct {
	Step      string
	PackageID string
}

var sessions = map[int64]*sessionState{}

// Main reply keyboard — always visible at the bottom
func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🛍 Shop"),
			tgbotapi.NewKeyboardButton("📦 My Orders"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💬 Support"),
			tgbotapi.NewKeyboardButton("📋 Rules"),
		),
	)
}

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

	case msg.Text == "🛍 Shop":
		b.sendPackageMenu(chatID)
		sess.Step = ""

	case msg.Text == "📦 My Orders":
		b.sendText(chatID, "📦 *Your Orders*\n\nNo active orders yet.\n\nTap 🛍 *Shop* to place your first order.")

	case msg.Text == "💬 Support":
		b.sendText(chatID, "💬 *Support*\n\nFor help with your order, contact us:\n👉 @AaPomSupport\n\nResponse time: within 1 hour.")

	case msg.Text == "📋 Rules":
		b.sendText(chatID, "📋 *Rules & Info*\n\n✅ Orders are non-refundable once placed\n✅ Delivery starts within 0–1 hours\n✅ Follower Booster includes 30-day refill\n✅ Use real public profile links\n\n⚠️ Private accounts will not be fulfilled")

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

		platformIcon := platformEmoji(string(pkg.Platform))
		b.sendText(chatID, fmt.Sprintf(
			"%s *%s*\n💰 KES %d\n📦 %s\n\n✏️ Paste your profile/post link below:",
			platformIcon, pkg.Name, pkg.PriceKES, pkg.Description,
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
		b.sendText(chatID, "⚠️ That doesn't look like a valid link.\n\nPlease paste the full URL, e.g:\n`https://instagram.com/yourprofile`")
		return
	}

	pkg, ok := GetPackage(sess.PackageID)
	if !ok {
		b.sendText(chatID, "Something went wrong. Tap 🛍 Shop to start over.")
		sess.Step = ""
		return
	}

	orderID, err := b.store.CreatePendingOrder(ctx, userID, pkg.ID, link, pkg.PriceKES)
	if err != nil {
		log.Printf("createPendingOrder: %v", err)
		b.sendText(chatID, "⚠️ Could not create your order. Please try again.")
		return
	}

	platformIcon := platformEmoji(string(pkg.Platform))
	b.sendTextWithKeyboard(chatID, fmt.Sprintf(
		"✅ *Order Received!*\n\n%s *%s*\n🔗 %s\n💰 *KES %d*\n\n📲 Send payment via M\\-Pesa to:\n*Till No: XXXXXX*\n\nShare the M\\-Pesa confirmation code with us once paid\\. Your order will be activated within minutes\\.",
		platformIcon, pkg.Name, link, pkg.PriceKES,
	), mainKeyboard())

	b.notifyAdmins(ctx, orderID, userID, pkg, link)
	sess.Step = ""
}

func (b *Bot) sendWelcome(chatID int64) {
	text := "👋 *Welcome to AaPom SMM!*\n\n🚀 Grow your social media fast and affordably\\.\n\n📱 We deliver real followers, likes \\& views for:\n• Instagram\n• TikTok\n• YouTube\n\nTap *🛍 Shop* to see available packages\\."
	b.sendTextWithKeyboard(chatID, text, mainKeyboard())
}

func (b *Bot) sendPackageMenu(chatID int64) {
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, pkg := range Catalog {
		icon := platformEmoji(string(pkg.Platform))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s — KES %d", icon, pkg.Name, pkg.PriceKES),
				"pkg:"+pkg.ID,
			),
		))
	}
	msg := tgbotapi.NewMessage(chatID, "🛍 *Choose a Package*\n\nAll packages include fast delivery & guaranteed quality:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ParseMode = "MarkdownV2"
	b.api.Send(msg)
}

func (b *Bot) sendBalance(ctx context.Context, chatID int64) {
	bal, err := b.wiz.GetBalance()
	if err != nil {
		b.sendText(chatID, fmt.Sprintf("⚠️ Balance check failed: %v", err))
		return
	}
	b.sendText(chatID, fmt.Sprintf("💰 *SMMWiz Balance*\n\n`%s %s`", bal.Balance, bal.Currency))
}

func (b *Bot) notifyAdmins(ctx context.Context, orderID, clientTelegramID int64, pkg Package, link string) {
	text := fmt.Sprintf(
		"🔔 *New Order \\#%d*\n\n👤 Client: `%d`\n📦 %s %s\n💰 KES %d\n🔗 %s\n\n✅ Tap *Approve* once M\\-Pesa payment is confirmed\\.",
		orderID, clientTelegramID,
		platformEmoji(string(pkg.Platform)), escapeMarkdown(pkg.Name),
		pkg.PriceKES, escapeMarkdown(link),
	)
	approveBtn := tgbotapi.NewInlineKeyboardButtonData("✅ Approve", fmt.Sprintf("approve:%d", orderID))
	rejectBtn := tgbotapi.NewInlineKeyboardButtonData("❌ Reject", fmt.Sprintf("reject:%d", orderID))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(approveBtn, rejectBtn))

	for _, adminID := range b.adminIDs {
		msg := tgbotapi.NewMessage(adminID, text)
		msg.ReplyMarkup = keyboard
		msg.ParseMode = "MarkdownV2"
		b.api.Send(msg)
	}
}

func (b *Bot) approveOrder(ctx context.Context, chatID, approverID int64, orderIDStr string, msgID int) {
	var orderID int64
	fmt.Sscanf(orderIDStr, "%d", &orderID)

	if err := b.store.ConfirmTransaction(ctx, orderID, approverID); err != nil {
		log.Printf("confirmTransaction: %v", err)
		b.sendText(chatID, "⚠️ Failed to confirm transaction.")
		return
	}

	go b.fulfillOrder(context.Background(), orderID)

	edit := tgbotapi.NewEditMessageText(chatID, msgID,
		fmt.Sprintf("✅ Order #%d approved — fulfillment started.", orderID))
	b.api.Send(edit)
}

func (b *Bot) rejectOrder(ctx context.Context, chatID, approverID int64, orderIDStr string, msgID int) {
	var orderID int64
	fmt.Sscanf(orderIDStr, "%d", &orderID)

	if err := b.store.CancelOrder(ctx, orderID); err != nil {
		log.Printf("cancelOrder: %v", err)
		return
	}

	edit := tgbotapi.NewEditMessageText(chatID, msgID,
		fmt.Sprintf("❌ Order #%d rejected.", orderID))
	b.api.Send(edit)
}

func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"
	b.api.Send(msg)
}

func (b *Bot) sendTextWithKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard
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

func platformEmoji(platform string) string {
	switch platform {
	case "tiktok":
		return "🎵"
	case "instagram":
		return "📸"
	case "youtube":
		return "▶️"
	default:
		return "📱"
	}
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
		">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	)
	return replacer.Replace(s)
}
