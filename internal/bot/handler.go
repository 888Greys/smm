package bot

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type sessionState struct {
	Step         string
	PackageID    string
	ProfileLink  string
	ReferralCode string
}

var sessions = map[int64]*sessionState{}

var phoneRegex = regexp.MustCompile(`^(?:254|\+254|0)(7\d{8}|1\d{8})$`)

func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🛍 Shop"),
			tgbotapi.NewKeyboardButton("📦 My Orders"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🤝 Refer & Earn"),
			tgbotapi.NewKeyboardButton("💬 Support"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
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

	log.Printf("msg from %d: %q", chatID, msg.Text)

	switch {
	case strings.HasPrefix(msg.Text, "/start"):
		b.store.UpsertClient(ctx, msg.From.ID)
		// Parse referral deep-link: /start ref_XXXXXXXX
		if parts := strings.Fields(msg.Text); len(parts) == 2 && strings.HasPrefix(parts[1], "ref_") {
			sess.ReferralCode = strings.TrimPrefix(parts[1], "ref_")
		}
		b.sendWelcome(chatID)
		sess.Step = ""

	case msg.Text == "🛍 Shop":
		b.sendPackageMenu(chatID)
		sess.Step = ""

	case msg.Text == "📦 My Orders":
		b.sendText(chatID, "📦 *My Orders*\n\nNo active orders yet.\n\nTap 🛍 *Shop* to place your first order.")

	case msg.Text == "🤝 Refer & Earn" || msg.Text == "/myreferral":
		b.sendReferralInfo(ctx, chatID, msg.From.ID)

	case msg.Text == "💬 Support":
		b.sendText(chatID, "💬 *Support*\n\nContact us: @AaPomSupport\n\nResponse time: within 1 hour.")

	case msg.Text == "📋 Rules":
		b.sendText(chatID, "📋 *Rules*\n\n✅ Orders are non-refundable once placed\n✅ Delivery starts within 0-1 hours\n✅ Packages with Refill include 30-day guarantee\n✅ Use real public profile links\n\n⚠️ Private accounts will not be fulfilled")

	case msg.Text == "/balance" && b.isAdmin(msg.From.ID):
		b.sendBalance(ctx, chatID)

	case msg.Text == "/stats" && b.isAdmin(msg.From.ID):
		b.sendStats(ctx, chatID)

	case sess.Step == "awaiting_link":
		b.handleLinkSubmission(ctx, chatID, msg.From.ID, msg.Text, sess)

	case sess.Step == "awaiting_phone":
		b.handlePhoneSubmission(ctx, chatID, msg.From.ID, msg.Text, sess)

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

		b.sendText(chatID, fmt.Sprintf(
			"%s *%s*\n💰 KES %d\n📦 %s\n\n✏️ Paste your profile/post link below:",
			platformEmoji(string(pkg.Platform)), pkg.Name, pkg.PriceKES, pkg.Description,
		))

	case strings.HasPrefix(cb.Data, "approve:"):
		if !b.isAdmin(cb.From.ID) {
			return
		}
		b.approveOrder(ctx, chatID, cb.From.ID, strings.TrimPrefix(cb.Data, "approve:"), cb.Message.MessageID)

	case strings.HasPrefix(cb.Data, "reject:"):
		if !b.isAdmin(cb.From.ID) {
			return
		}
		b.rejectOrder(ctx, chatID, cb.From.ID, strings.TrimPrefix(cb.Data, "reject:"), cb.Message.MessageID)
	}
}

func (b *Bot) handleLinkSubmission(ctx context.Context, chatID, userID int64, link string, sess *sessionState) {
	link = strings.TrimSpace(link)
	if !isValidLink(link) {
		b.sendText(chatID, "⚠️ That doesn't look like a valid link.\n\nPlease paste the full URL, e.g:\nhttps://instagram.com/yourprofile")
		return
	}
	sess.ProfileLink = link
	sess.Step = "awaiting_phone"
	b.sendText(chatID, "📲 Enter your M-Pesa phone number to pay:\n\nFormat: *07XXXXXXXX* or *254XXXXXXXXX*")
}

func (b *Bot) handlePhoneSubmission(ctx context.Context, chatID, userID int64, phone string, sess *sessionState) {
	phone = strings.TrimSpace(phone)
	normalized := normalizePhone(phone)
	if normalized == "" {
		b.sendText(chatID, "⚠️ Invalid phone number. Please enter a valid Safaricom number, e.g:\n*0712345678*")
		return
	}

	pkg, ok := GetPackage(sess.PackageID)
	if !ok {
		b.sendText(chatID, "Something went wrong. Tap 🛍 Shop to start over.")
		sess.Step = ""
		return
	}

	orderID, err := b.store.CreatePendingOrder(ctx, userID, pkg.ID, sess.ProfileLink, pkg.PriceKES, sess.ReferralCode)
	if err != nil {
		log.Printf("createPendingOrder: %v", err)
		b.sendText(chatID, "⚠️ Could not create your order. Please try again.")
		return
	}

	b.sendText(chatID, fmt.Sprintf(
		"💳 Sending M-Pesa request to *%s*...\n\nCheck your phone and enter your PIN to complete payment.",
		phone,
	))

	go b.initiatePayment(context.Background(), chatID, orderID, pkg.PriceKES, normalized, phone)
	sess.Step = ""
}

func (b *Bot) initiatePayment(ctx context.Context, chatID, orderID int64, amountKES int, phone, displayPhone string) {
	reference := fmt.Sprintf("Order #%d", orderID)
	resp, err := b.pay.InitiateSTK(amountKES, phone, reference)
	if err != nil {
		log.Printf("initiateSTK order %d: %v", orderID, err)
		b.sendText(chatID, "⚠️ Could not send M-Pesa request. Please try again or contact support.")
		return
	}

	if err := b.store.SaveSTKRequest(ctx, orderID, phone, resp.TransactionRequestID); err != nil {
		log.Printf("saveSTKRequest order %d: %v", orderID, err)
	}

	log.Printf("STK push sent: order %d phone %s txn %s", orderID, displayPhone, resp.TransactionRequestID)
}

func (b *Bot) sendWelcome(chatID int64) {
	b.sendTextWithKeyboard(chatID,
		"👋 *Welcome to AaPom SMM!*\n\n🚀 Grow your social media fast and affordably.\n\n📱 We deliver real followers, likes & views for:\n• Instagram\n• TikTok\n• YouTube\n\nTap 🛍 *Shop* to see available packages.",
		mainKeyboard(),
	)
}

func (b *Bot) sendPackageMenu(chatID int64) {
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, pkg := range Catalog {
		if pkg.ID == "test_ksh1" {
			continue // hide test package from public menu
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s — KES %d", platformEmoji(string(pkg.Platform)), pkg.Name, pkg.PriceKES),
				"pkg:"+pkg.ID,
			),
		))
	}
	msg := tgbotapi.NewMessage(chatID, "🛍 *Choose a Package*\n\nAll packages include fast delivery & guaranteed quality:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ParseMode = "Markdown"
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("sendPackageMenu: %v", err)
	}
}

func (b *Bot) sendBalance(ctx context.Context, chatID int64) {
	bal, err := b.wiz.GetBalance()
	if err != nil {
		b.sendText(chatID, fmt.Sprintf("⚠️ Balance check failed: %v", err))
		return
	}
	b.sendText(chatID, fmt.Sprintf("💰 *SMMWiz Balance*\n\n%s %s", bal.Balance, bal.Currency))
}

func (b *Bot) sendStats(ctx context.Context, chatID int64) {
	st, err := b.store.GetStats(ctx)
	if err != nil {
		b.sendText(chatID, fmt.Sprintf("⚠️ Stats error: %v", err))
		return
	}

	var totalRevenue, totalProfit, totalOrders int
	lines := ""
	for _, line := range st.Lines {
		pkg, ok := GetPackage(line.PackageID)
		profitLine := 0
		name := line.PackageID
		if ok {
			profitLine = pkg.MarginKES * line.OrderCount
			name = pkg.Name
		}
		totalRevenue += line.RevenueKES
		totalProfit += profitLine
		totalOrders += line.OrderCount
		lines += fmt.Sprintf("  • %s × %d → KES %d\n", name, line.OrderCount, line.RevenueKES)
	}
	if lines == "" {
		lines = "  No paid orders yet today.\n"
	}

	bal, _ := b.wiz.GetBalance()
	wizBal := "unknown"
	if bal != nil {
		wizBal = bal.Balance + " " + bal.Currency
	}

	text := fmt.Sprintf(
		"📊 *Admin Dashboard — Last 24h*\n\n"+
			"💰 Revenue: *KES %d* (%d orders)\n"+
			"📈 Est. Profit: *KES %d*\n\n"+
			"%s\n"+
			"📋 *All-time:* %d total | %d pending | %d processing | %d completed\n\n"+
			"🏦 SMMWiz: *%s*",
		totalRevenue, totalOrders, totalProfit,
		lines,
		st.TotalOrders, st.PendingOrders, st.ProcessingOrders, st.CompletedOrders,
		wizBal,
	)
	b.sendText(chatID, text)
}

func (b *Bot) sendReferralInfo(ctx context.Context, chatID, telegramID int64) {
	b.store.UpsertClient(ctx, telegramID)

	code, err := b.store.GetOrCreateReferralCode(ctx, telegramID)
	if err != nil {
		log.Printf("getReferralCode %d: %v", telegramID, err)
		b.sendText(chatID, "⚠️ Could not load referral info. Try again.")
		return
	}

	balance, _ := b.store.GetCreditBalance(ctx, telegramID)
	link := fmt.Sprintf("https://t.me/%s?start=ref_%s", b.api.Self.UserName, code)

	b.sendText(chatID, fmt.Sprintf(
		"🤝 *Refer & Earn*\n\n"+
			"Share your link and earn *KES 50* every time a friend places their first order!\n\n"+
			"🔗 Your link:\n`%s`\n\n"+
			"💳 Your credit balance: *KES %d*\n\n"+
			"_Credits can be used toward your next order — contact support to redeem._",
		link, balance,
	))
}

func (b *Bot) notifyAdmins(ctx context.Context, orderID, clientTelegramID int64, pkg Package, link string) {
	text := fmt.Sprintf(
		"🔔 *New Order #%d*\n\n👤 Client: %d\n📦 %s %s\n💰 KES %d\n🔗 %s",
		orderID, clientTelegramID, platformEmoji(string(pkg.Platform)), pkg.Name, pkg.PriceKES, link,
	)
	for _, adminID := range b.adminIDs {
		msg := tgbotapi.NewMessage(adminID, text)
		msg.ParseMode = "Markdown"
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

	edit := tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("✅ Order #%d approved — fulfillment started.", orderID))
	b.api.Send(edit)
}

func (b *Bot) rejectOrder(ctx context.Context, chatID, approverID int64, orderIDStr string, msgID int) {
	var orderID int64
	fmt.Sscanf(orderIDStr, "%d", &orderID)

	if err := b.store.CancelOrder(ctx, orderID); err != nil {
		log.Printf("cancelOrder: %v", err)
		return
	}

	edit := tgbotapi.NewEditMessageText(chatID, msgID, fmt.Sprintf("❌ Order #%d rejected.", orderID))
	b.api.Send(edit)
}

func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("sendText to %d: %v", chatID, err)
	}
}

func (b *Bot) sendTextWithKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("sendTextWithKeyboard to %d: %v", chatID, err)
	}
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

// normalizePhone converts Kenyan numbers to 254XXXXXXXXX format
func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	if !phoneRegex.MatchString(phone) {
		return ""
	}
	if strings.HasPrefix(phone, "0") {
		return "254" + phone[1:]
	}
	if strings.HasPrefix(phone, "+") {
		return phone[1:]
	}
	return phone
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
