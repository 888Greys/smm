package bot

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/aapom/smm/internal/profile"
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
			tgbotapi.NewKeyboardButton("❓ How it Works"),
			tgbotapi.NewKeyboardButton("🛡️ Account Safety"),
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
		isNew, _ := b.store.UpsertClient(ctx, msg.From.ID)
		if parts := strings.Fields(msg.Text); len(parts) == 2 && strings.HasPrefix(parts[1], "ref_") {
			sess.ReferralCode = strings.TrimPrefix(parts[1], "ref_")
		}
		if isNew {
			b.notifier.NotifyNewUser(msg.From.ID, msg.From.UserName, msg.From.FirstName, sess.ReferralCode)
		} else {
			b.notifier.NotifyReturningUser(msg.From.ID, msg.From.UserName, msg.From.FirstName)
		}
		b.sendWelcome(chatID)
		sess.Step = ""

	case msg.Text == "🛍 Shop":
		b.sendCategoryMenu(chatID)
		sess.Step = ""

	case msg.Text == "📦 My Orders":
		b.sendText(chatID, "📦 *My Orders*\n\nNo active orders yet.\n\nTap 🛍 *Shop* to place your first order.")

	case msg.Text == "❓ How it Works":
		b.sendHowItWorks(chatID)

	case msg.Text == "🛡️ Account Safety":
		b.sendAccountSafety(chatID)

	case msg.Text == "🤝 Refer & Earn" || msg.Text == "/myreferral":
		b.sendReferralInfo(ctx, chatID, msg.From.ID)

	case msg.Text == "💬 Support":
		b.sendText(chatID, "💬 *Support*\n\nDM us directly: @workratew\n\n⏱ Response time: within 1 hour.\n\n_For order issues include your Order # in the message._")

	case msg.Text == "/balance" && b.isAdmin(msg.From.ID):
		b.sendBalance(ctx, chatID)

	case msg.Text == "/stats" && b.isAdmin(msg.From.ID):
		b.sendStats(ctx, chatID)

	case sess.Step == "awaiting_link":
		b.handleLinkSubmission(ctx, chatID, msg.From.ID, msg.Text, sess)

	case sess.Step == "awaiting_confirm":
		b.sendText(chatID, "⬆️ Please tap the *Confirm & Pay* button above to proceed.")

	case sess.Step == "awaiting_phone":
		b.handlePhoneSubmission(ctx, chatID, msg.From.ID, msg.Text, sess)

	default:
		b.sendCategoryMenu(chatID)
	}
}

func (b *Bot) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	sess := b.getSession(chatID)

	b.api.Send(tgbotapi.NewCallback(cb.ID, ""))

	switch {
	case strings.HasPrefix(cb.Data, "cat:"):
		cat := strings.TrimPrefix(cb.Data, "cat:")
		if cat == "back" {
			b.editToCategoryMenu(chatID, cb.Message.MessageID)
		} else {
			b.editToPlatformPackages(chatID, cb.Message.MessageID, cat)
		}

	case strings.HasPrefix(cb.Data, "pkg:"):
		pkgID := strings.TrimPrefix(cb.Data, "pkg:")
		pkg, ok := GetPackage(pkgID)
		if !ok {
			b.sendText(chatID, "Package not found.")
			return
		}
		sess.PackageID = pkgID
		sess.Step = "awaiting_link"
		b.notifier.NotifyPackageSelected(cb.From.ID, cb.From.UserName, pkg)

		b.sendText(chatID, fmt.Sprintf(
			"%s *%s*\n\n📦 %s\n💰 Price: *KES %d*\n%s\n\n✏️ *Step 1 of 4* — Enter your %s username:\n\n_Just your username, e.g._ `yourhandle` _(no @ needed)_\n_Make sure your profile is Public_",
			platformEmoji(string(pkg.Platform)),
			pkg.Name,
			pkg.Description,
			pkg.PriceKES,
			refillLine(pkg),
			platformName(string(pkg.Platform)),
		))

	case cb.Data == "confirm_order":
		if sess.Step != "awaiting_confirm" || sess.PackageID == "" {
			b.api.Send(tgbotapi.NewCallback(cb.ID, "Session expired — tap Shop to restart."))
			return
		}
		// Lock in the confirmation visually
		edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID,
			cb.Message.Text+"\n\n✅ _Order confirmed — proceeding to payment..._")
		edit.ParseMode = "Markdown"
		b.api.Send(edit)

		sess.Step = "awaiting_phone"
		b.sendText(chatID, "📲 *Step 3 of 4* — Enter your M-Pesa number:\n\nFormat: `07XXXXXXXX` or `254XXXXXXXXX`")

	case cb.Data == "change_link":
		if sess.PackageID == "" {
			b.sendText(chatID, "Session expired. Tap 🛍 Shop to restart.")
			return
		}
		sess.Step = "awaiting_link"
		sess.ProfileLink = ""
		pkg, _ := GetPackage(sess.PackageID)
		edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID,
			"✏️ No problem — paste your correct "+platformName(string(pkg.Platform))+" profile link:")
		edit.ParseMode = "Markdown"
		b.api.Send(edit)

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

// ── Navigation ───────────────────────────────────────────────────────────────

func (b *Bot) sendCategoryMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "🛍 *SMM Mall*\n\nChoose a platform to see available packages:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = categoryKeyboard()
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("sendCategoryMenu: %v", err)
	}
}

func (b *Bot) editToCategoryMenu(chatID int64, msgID int) {
	edit := tgbotapi.NewEditMessageText(chatID, msgID, "🛍 *SMM Mall*\n\nChoose a platform to see available packages:")
	edit.ParseMode = "Markdown"
	kb := categoryKeyboard()
	edit.ReplyMarkup = &kb
	b.api.Send(edit)
}

func (b *Bot) editToPlatformPackages(chatID int64, msgID int, category string) {
	packages := CategoryPackages(category)

	header := categoryHeader(category)
	if len(packages) == 0 {
		edit := tgbotapi.NewEditMessageText(chatID, msgID, header+"\n\nNo packages available yet.")
		edit.ParseMode = "Markdown"
		b.api.Send(edit)
		return
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, pkg := range packages {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s — KES %d", platformEmoji(string(pkg.Platform)), pkg.Name, pkg.PriceKES),
				"pkg:"+pkg.ID,
			),
		))
	}
	// Back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Back to Categories", "cat:back"),
	))

	edit := tgbotapi.NewEditMessageText(chatID, msgID, header)
	edit.ParseMode = "Markdown"
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	edit.ReplyMarkup = &kb
	b.api.Send(edit)
}

func categoryKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎵 TikTok", "cat:tiktok"),
			tgbotapi.NewInlineKeyboardButtonData("📸 Instagram", "cat:instagram"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ YouTube", "cat:youtube"),
			tgbotapi.NewInlineKeyboardButtonData("💎 Combo Deals", "cat:combo"),
		),
	)
}

func categoryHeader(cat string) string {
	headers := map[string]string{
		"tiktok":    "🎵 *TikTok Packages*\n\nGrow your TikTok with real followers, views & likes:",
		"instagram": "📸 *Instagram Packages*\n\nBoost your Instagram with HQ followers and engagement:",
		"youtube":   "▶️ *YouTube Packages*\n\nGrow your channel with real subscribers and views:",
		"combo":     "💎 *Combo Deals*\n\nMaximum growth at the best value — our highest-margin packages:",
	}
	if h, ok := headers[cat]; ok {
		return h
	}
	return "📦 *Packages*"
}

// ── Order flow ────────────────────────────────────────────────────────────────

func (b *Bot) handleLinkSubmission(ctx context.Context, chatID, userID int64, input string, sess *sessionState) {
	input = strings.TrimSpace(input)

	pkg, ok := GetPackage(sess.PackageID)
	if !ok {
		b.sendText(chatID, "Something went wrong. Tap 🛍 Shop to start over.")
		sess.Step = ""
		return
	}

	var profileLink string
	if isValidLink(input) {
		// User pasted a full URL — accept it directly
		profileLink = input
	} else {
		// Treat as a username — construct the URL
		username := strings.TrimPrefix(input, "@")
		if len(username) < 2 || strings.ContainsAny(username, " /\\?#") {
			b.sendText(chatID, "⚠️ That doesn't look right.\n\nPlease enter just your username, e.g:\n`yourhandle`")
			return
		}
		profileLink = profile.ProfileURL(string(pkg.Platform), username)
	}

	sess.ProfileLink = profileLink
	sess.Step = "awaiting_confirm"
	b.sendSafetyBriefing(chatID, pkg, profileLink)
}

func (b *Bot) sendSafetyBriefing(chatID int64, pkg Package, link string) {
	followers := totalFollowersInPackage(pkg)

	// Truncate long links for display
	displayLink := link
	if len(displayLink) > 45 {
		displayLink = displayLink[:42] + "..."
	}

	var safetyBlock string
	if followers > 1000 {
		// ~400/day target, capped between 200 and 500 for display
		dailyHigh := followers / (followers / 400)
		if dailyHigh < 200 {
			dailyHigh = 200
		}
		if dailyHigh > 500 {
			dailyHigh = 500
		}
		dailyLow := dailyHigh - 100
		safetyBlock = fmt.Sprintf(
			"🚀 *Safety Protocol Active*\n"+
				"Your *%s followers* will be drip-fed at ~*%d–%d/day* to perfectly mimic organic growth and keep your account safe from platform flags.\n",
			formatCount(followers), dailyLow, dailyHigh,
		)
	} else if followers > 0 {
		safetyBlock = fmt.Sprintf(
			"⚡ *Fast Delivery*\nYour *%s followers* will start arriving within minutes of payment.\n",
			formatCount(followers),
		)
	} else {
		safetyBlock = "⚡ *Fast Delivery*\nYour order will start processing immediately after payment.\n"
	}

	refill := ""
	if pkg.Refillable {
		refill = "\n🔄 *30-Day Refill Guarantee* included — if followers drop, we top them back up automatically."
	}

	text := fmt.Sprintf(
		"✅ *Order Prepared — Step 2 of 4*\n\n"+
			"Here's exactly what will happen:\n\n"+
			"📦 *%s*\n"+
			"🔗 Profile: `%s`\n"+
			"💰 Total: *KES %d*\n\n"+
			"%s"+
			"%s\n\n"+
			"_Everything looks correct? Tap the button below to proceed to M-Pesa payment._",
		pkg.Name,
		displayLink,
		pkg.PriceKES,
		safetyBlock,
		refill,
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Confirm & Pay via M-Pesa", "confirm_order"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Change Profile Link", "change_link"),
		),
	)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("sendSafetyBriefing %d: %v", chatID, err)
	}
}

func totalFollowersInPackage(pkg Package) int {
	followerServices := map[int]bool{18612: true, 20888: true}
	total := 0
	for _, comp := range pkg.Components {
		if followerServices[comp.ServiceID] {
			total += comp.Quantity
		}
	}
	return total
}

func formatCount(n int) string {
	if n >= 1000 {
		if n%1000 == 0 {
			return fmt.Sprintf("%d,000", n/1000)
		}
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
}

func (b *Bot) handlePhoneSubmission(ctx context.Context, chatID, userID int64, phone string, sess *sessionState) {
	phone = strings.TrimSpace(phone)
	normalized := normalizePhone(phone)
	if normalized == "" {
		b.sendText(chatID, "⚠️ Invalid number. Please enter a valid Safaricom number:\n`0712345678`")
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
		"💳 *Step 4 of 4* — M-Pesa request sent to `%s`\n\n📱 Check your phone and enter your *M-Pesa PIN* to complete payment.\n\n_Order #%d · KES %d_",
		phone, orderID, pkg.PriceKES,
	))

	go b.initiatePayment(context.Background(), chatID, orderID, pkg.PriceKES, normalized, phone)
	b.notifier.NotifyOrderCreated(orderID, userID, "", pkg, phone)
	sess.Step = ""
}

func (b *Bot) initiatePayment(ctx context.Context, chatID, orderID int64, amountKES int, phone, displayPhone string) {
	reference := fmt.Sprintf("Order #%d", orderID)
	resp, err := b.pay.InitiateSTK(amountKES, phone, reference)
	if err != nil {
		log.Printf("initiateSTK order %d: %v", orderID, err)
		b.sendText(chatID, "⚠️ Could not send M-Pesa request. Please try again or contact @workratew.")
		return
	}

	if err := b.store.SaveSTKRequest(ctx, orderID, phone, resp.TransactionRequestID); err != nil {
		log.Printf("saveSTKRequest order %d: %v", orderID, err)
	}

	log.Printf("STK push sent: order %d phone %s txn %s", orderID, displayPhone, resp.TransactionRequestID)
}

// ── Info pages ────────────────────────────────────────────────────────────────

func (b *Bot) sendWelcome(chatID int64) {
	b.sendTextWithKeyboard(chatID,
		"👋 *Welcome to VectorBoost!*\n\n🚀 Grow your social media fast & affordably.\n\n"+
			"✅ Real followers, likes & views\n"+
			"✅ TikTok • Instagram • YouTube\n"+
			"✅ Pay securely with M-Pesa\n"+
			"✅ Delivery starts within the hour\n\n"+
			"Tap 🛍 *Shop* to browse packages.",
		mainKeyboard(),
	)
}

func (b *Bot) sendHowItWorks(chatID int64) {
	b.sendText(chatID,
		"❓ *How It Works*\n\n"+
			"*1️⃣ Choose a Package*\n"+
			"Tap 🛍 Shop, select your platform, then pick the package that fits your goal and budget.\n\n"+
			"*2️⃣ Paste Your Profile Link*\n"+
			"Send your public TikTok, Instagram or YouTube profile URL. Make sure your account is public.\n\n"+
			"*3️⃣ Pay with M-Pesa*\n"+
			"Enter your Safaricom number. You'll receive an STK push on your phone — enter your PIN to confirm.\n\n"+
			"*4️⃣ Delivery Starts Automatically*\n"+
			"Our system places your order instantly. Followers and views start arriving within minutes to a few hours.\n\n"+
			"*5️⃣ Get Notified When Done*\n"+
			"You'll receive a message here when delivery is complete. Packages marked 🔄 include a 30-day refill guarantee.\n\n"+
			"⏱ *Average delivery: Under 1 hour*\n\n"+
			"Questions? Contact @workratew",
	)
}

func (b *Bot) sendAccountSafety(chatID int64) {
	b.sendText(chatID,
		"🛡️ *Account Safety — Anti-Ban Protection*\n\n"+
			"Your account safety is our #1 priority. Here's how we protect you:\n\n"+
			"✅ *Organic-Speed Delivery*\n"+
			"We never dump thousands of followers at once. Our system delivers at a pace that looks 100% natural to Instagram and TikTok's algorithms.\n\n"+
			"✅ *Drip-Feed Technology*\n"+
			"Large orders are split into smaller daily batches:\n"+
			"• Followers are added at a safe rate of ~200–500 per day\n"+
			"• This mirrors the growth pattern of viral organic content\n"+
			"• No sudden spikes that trigger spam detection\n\n"+
			"✅ *High-Quality Accounts*\n"+
			"All our services use real-looking, aged accounts — not obvious bots with no posts or profile pictures.\n\n"+
			"✅ *30-Day Refill Guarantee*\n"+
			"If any followers drop within 30 days, our system tops them back up automatically at no extra cost.\n\n"+
			"⚠️ *Best Practices for You:*\n"+
			"• Keep your profile *public* during delivery\n"+
			"• Don't change your username during an active order\n"+
			"• Avoid buying from multiple providers at the same time\n\n"+
			"💬 Still have questions? DM @workratew",
	)
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
		lines += fmt.Sprintf("  • %s ×%d → KES %d\n", name, line.OrderCount, line.RevenueKES)
	}
	if lines == "" {
		lines = "  No paid orders yet today.\n"
	}

	bal, _ := b.wiz.GetBalance()
	wizBal := "unknown"
	if bal != nil {
		wizBal = bal.Balance + " " + bal.Currency
	}

	b.sendText(chatID, fmt.Sprintf(
		"📊 *Admin Dashboard — Last 24h*\n\n"+
			"💰 Revenue: *KES %d* (%d orders)\n"+
			"📈 Est. Profit: *KES %d*\n\n"+
			"%s\n"+
			"📋 *All-time:* %d total | %d pending | %d processing | %d completed\n\n"+
			"🏦 SMMWiz balance: *%s*",
		totalRevenue, totalOrders, totalProfit,
		lines,
		st.TotalOrders, st.PendingOrders, st.ProcessingOrders, st.CompletedOrders,
		wizBal,
	))
}

func (b *Bot) sendReferralInfo(ctx context.Context, chatID, telegramID int64) {
	b.store.UpsertClient(ctx, telegramID) //nolint:errcheck

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
			"🔗 *Your referral link:*\n`%s`\n\n"+
			"💳 Your credit balance: *KES %d*\n\n"+
			"_Credits can be used toward your next order. DM @workratew to redeem._",
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

// ── Helpers ───────────────────────────────────────────────────────────────────

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

func platformName(platform string) string {
	switch platform {
	case "tiktok":
		return "TikTok"
	case "instagram":
		return "Instagram"
	case "youtube":
		return "YouTube"
	default:
		return "social media"
	}
}

func refillLine(pkg Package) string {
	if pkg.Refillable {
		return "🔄 30-day refill guarantee"
	}
	return ""
}
