package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/aapom/smm/internal/bot"
	"github.com/aapom/smm/internal/db"
	"github.com/aapom/smm/internal/megapay"
	"github.com/aapom/smm/internal/profile"
)

func main() {
	godotenv.Load()

	store, err := db.NewStore(context.Background(), mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer store.Close()

	pay := megapay.New(mustEnv("MEGAPAY_API_KEY"), mustEnv("MEGAPAY_EMAIL"))
	webhookSecret := os.Getenv("MEGAPAY_WEBHOOK_SECRET")
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN") // e.g. https://innbucks.org

	// Telegram notifier:
	//   client messages  → main bot token  (user's chat is with the main bot)
	//   admin messages   → admin bot token + ADMIN_CHAT_ID  (all admin traffic in one place)
	adminChatID, _ := strconv.ParseInt(os.Getenv("ADMIN_CHAT_ID"), 10, 64)
	tg := newTGNotifier(
		os.Getenv("TELEGRAM_BOT_TOKEN"), // used to message clients
		os.Getenv("ADMIN_BOT_TOKEN"),    // used to message admin
		adminChatID,
	)

	mux := http.NewServeMux()

	// Webhooks
	mux.HandleFunc("/webhook/megapay", megapayHandler(store, tg, webhookSecret))

	// REST API
	mux.HandleFunc("/api/packages", packagesHandler())
	mux.HandleFunc("/api/orders", ordersHandler(store, pay))
	mux.HandleFunc("/api/orders/", orderStatusHandler(store)) // /api/orders/:id
	mux.HandleFunc("/api/profile", profileLookupHandler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := corsMiddleware(mux, frontendOrigin)

	srv := &http.Server{
		Addr:         ":8005",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("API server listening on :8005")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Println("server stopped")
}

// ── CORS ─────────────────────────────────────────────────────────────────────

func corsMiddleware(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── REST API handlers ────────────────────────────────────────────────────────

type packageDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	PriceKES    int    `json:"price_kes"`
	Description string `json:"description"`
}

func packagesHandler() http.HandlerFunc {
	// Build once at startup
	var pkgs []packageDTO
	for _, p := range bot.Catalog {
		if p.ID == "test_ksh1" {
			continue
		}
		pkgs = append(pkgs, packageDTO{
			ID:          p.ID,
			Name:        p.Name,
			Platform:    string(p.Platform),
			PriceKES:    p.PriceKES,
			Description: p.Description,
		})
	}
	data, _ := json.Marshal(pkgs)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

type createOrderReq struct {
	PackageID    string `json:"package_id"`
	ProfileLink  string `json:"profile_link"`
	Phone        string `json:"phone"`
	ReferralCode string `json:"referral_code"`
}

type createOrderResp struct {
	OrderID int64  `json:"order_id"`
	Message string `json:"message"`
}

func ordersHandler(store *db.Store, pay *megapay.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req createOrderReq
		if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
			jsonError(w, "invalid request", http.StatusBadRequest)
			return
		}

		pkg, ok := bot.GetPackage(req.PackageID)
		if !ok || req.PackageID == "test_ksh1" {
			jsonError(w, "unknown package", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(req.ProfileLink, "http") {
			jsonError(w, "invalid profile link", http.StatusBadRequest)
			return
		}
		phone := normalizePhone(req.Phone)
		if phone == "" {
			jsonError(w, "invalid phone number", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// Web orders use telegram_id=0 as a sentinel (no Telegram account)
		orderID, err := store.CreatePendingOrder(ctx, 0, pkg.ID, req.ProfileLink, pkg.PriceKES, req.ReferralCode)
		if err != nil {
			log.Printf("web createOrder: %v", err)
			jsonError(w, "could not create order", http.StatusInternalServerError)
			return
		}

		reference := fmt.Sprintf("Order #%d", orderID)
		stkResp, err := pay.InitiateSTK(pkg.PriceKES, phone, reference)
		if err != nil {
			log.Printf("web STK order %d: %v", orderID, err)
			jsonError(w, "could not send M-Pesa request: "+err.Error(), http.StatusBadGateway)
			return
		}

		if err := store.SaveSTKRequest(ctx, orderID, phone, stkResp.TransactionRequestID); err != nil {
			log.Printf("web saveSTK order %d: %v", orderID, err)
		}

		log.Printf("web order %d created: pkg=%s phone=%s txn=%s", orderID, pkg.ID, phone, stkResp.TransactionRequestID)

		jsonOK(w, createOrderResp{
			OrderID: orderID,
			Message: fmt.Sprintf("M-Pesa request sent to %s. Enter your PIN to confirm.", req.Phone),
		})
	}
}

type orderStatusResp struct {
	OrderID     int64  `json:"order_id"`
	Status      string `json:"status"`
	PackageName string `json:"package_name"`
	Platform    string `json:"platform"`
	Description string `json:"description"`
	PriceKES    int    `json:"price_kes"`
	CreatedAt   string `json:"created_at"`
}

func orderStatusHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract ID from /api/orders/123
		idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
		var orderID int64
		fmt.Sscanf(idStr, "%d", &orderID)
		if orderID == 0 {
			jsonError(w, "invalid order id", http.StatusBadRequest)
			return
		}

		order, err := store.GetOrder(r.Context(), orderID)
		if err != nil {
			jsonError(w, "order not found", http.StatusNotFound)
			return
		}

		pkg, _ := bot.GetPackage(order.PackageID)

		jsonOK(w, orderStatusResp{
			OrderID:     order.ID,
			Status:      string(order.Status),
			PackageName: pkg.Name,
			Platform:    string(pkg.Platform),
			Description: pkg.Description,
			PriceKES:    order.TotalKES,
			CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		})
	}
}

// ── Profile lookup ────────────────────────────────────────────────────────────

func profileLookupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		platform := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("platform")))
		username := strings.TrimSpace(r.URL.Query().Get("username"))
		if platform == "" || username == "" {
			jsonError(w, "platform and username required", http.StatusBadRequest)
			return
		}
		info, err := profile.Lookup(platform, username)
		if err != nil {
			jsonError(w, "lookup failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, info)
	}
}

// ── MegaPay webhook ──────────────────────────────────────────────────────────

type megapayPayload struct {
	Reference string  `json:"reference"`
	OrderID   int64   `json:"order_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
	MpesaRef  string  `json:"mpesa_ref"`
	Signature string  `json:"signature"`
}

func megapayHandler(store *db.Store, tg *tgNotifier, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		if secret != "" {
			sig := r.Header.Get("X-MegaPay-Signature")
			if !validSignature(body, sig, secret) {
				log.Printf("megapay webhook: invalid signature")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		var payload megapayPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}

		log.Printf("megapay webhook: order %d status=%s ref=%s", payload.OrderID, payload.Status, payload.MpesaRef)

		if payload.Status != "SUCCESS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		ctx := r.Context()
		if err := store.ConfirmTransaction(ctx, payload.OrderID, 0); err != nil {
			log.Printf("megapay confirmTransaction order %d: %v", payload.OrderID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Notify client + admins in background — do NOT auto-fulfill
		go notifyPaymentConfirmed(context.Background(), store, tg, payload.OrderID, payload.MpesaRef)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}
}

func notifyPaymentConfirmed(ctx context.Context, store *db.Store, tg *tgNotifier, orderID int64, mpesaRef string) {
	order, err := store.GetOrder(ctx, orderID)
	if err != nil {
		log.Printf("notifyPaymentConfirmed getOrder %d: %v", orderID, err)
		return
	}

	pkg, _ := bot.GetPackage(order.PackageID)
	clientTgID, _ := store.GetClientTelegramID(ctx, orderID)

	// Tell the client their payment is in (via main bot — that's where they chatted)
	if clientTgID > 0 {
		tg.sendClient(clientTgID, fmt.Sprintf(
			"✅ *Payment Confirmed!*\n\n"+
				"We've received your M-Pesa payment for *%s* (KES %d).\n\n"+
				"🔧 Our team is reviewing your order and will kick off your boost shortly.\n"+
				"You'll get a message here the moment delivery starts.\n\n"+
				"_Order #%d · Ref: %s_",
			pkg.Name, order.TotalKES, orderID, mpesaRef,
		))
	}

	// Alert admin via admin bot with Fulfill / Reject buttons
	profileDisplay := order.ProfileLink
	if len(profileDisplay) > 50 {
		profileDisplay = profileDisplay[:47] + "..."
	}
	tg.sendAdmin(
		fmt.Sprintf(
			"💰 *Payment Confirmed — Order #%d*\n\n"+
				"📦 %s\n"+
				"💰 KES %d\n"+
				"🔗 %s\n"+
				"👤 Client TG: `%d`\n"+
				"📱 M-Pesa ref: `%s`\n"+
				"🕐 %s\n\n"+
				"Tap *Fulfill* to send to SMMWiz.",
			orderID, pkg.Name, order.TotalKES,
			profileDisplay, clientTgID, mpesaRef,
			time.Now().Format("02 Jan 15:04 MST"),
		),
		&tgInlineKb{
			InlineKeyboard: [][]tgInlineBtn{
				{
					{Text: "✅ Fulfill Order", CallbackData: fmt.Sprintf("fulfill:%d", orderID)},
					{Text: "❌ Reject", CallbackData: fmt.Sprintf("reject:%d", orderID)},
				},
			},
		},
	)
	log.Printf("order %d payment confirmed, admin notified via admin bot", orderID)
}

// ── Telegram notifier (lightweight — no library, just HTTP) ──────────────────

type tgNotifier struct {
	mainToken  string // main bot — used to message clients
	adminToken string // admin bot — used to message admin chat
	adminChat  int64
}

type tgInlineKb struct {
	InlineKeyboard [][]tgInlineBtn `json:"inline_keyboard"`
}

type tgInlineBtn struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func newTGNotifier(mainToken, adminToken string, adminChat int64) *tgNotifier {
	if adminToken == "" {
		log.Println("ADMIN_BOT_TOKEN not set — admin payment notifications disabled")
	}
	return &tgNotifier{mainToken: mainToken, adminToken: adminToken, adminChat: adminChat}
}

func (n *tgNotifier) sendClient(chatID int64, text string) {
	n.post(n.mainToken, chatID, text, nil)
}

func (n *tgNotifier) sendAdmin(text string, kb *tgInlineKb) {
	if n.adminToken == "" || n.adminChat == 0 {
		return
	}
	n.post(n.adminToken, n.adminChat, text, kb)
}

func (n *tgNotifier) post(token string, chatID int64, text string, kb *tgInlineKb) {
	if token == "" {
		return
	}
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	if kb != nil {
		payload["reply_markup"] = kb
	}
	body, _ := json.Marshal(payload)
	url := "https://api.telegram.org/bot" + token + "/sendMessage"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("tgNotifier post to %d: %v", chatID, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("tgNotifier post to %d: status %d: %s", chatID, resp.StatusCode, b)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func validSignature(body []byte, sig, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	if len(phone) == 10 && strings.HasPrefix(phone, "0") {
		return "254" + phone[1:]
	}
	if len(phone) == 12 && strings.HasPrefix(phone, "254") {
		return phone
	}
	if len(phone) == 13 && strings.HasPrefix(phone, "+254") {
		return phone[1:]
	}
	return ""
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
