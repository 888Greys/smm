package main

import (
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
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/aapom/smm/internal/db"
	"github.com/aapom/smm/internal/smmwiz"
	"github.com/aapom/smm/internal/bot"
)

func main() {
	godotenv.Load()

	store, err := db.NewStore(context.Background(), mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer store.Close()

	wiz := smmwiz.New(mustEnv("SMMWIZ_API_KEY"))
	webhookSecret := os.Getenv("MEGAPAY_WEBHOOK_SECRET")

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/megapay", megapayHandler(store, wiz, webhookSecret))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         ":8005",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("webhook server listening on :8005")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Println("webhook server stopped")
}

// MegaPay webhook payload — adjust fields to match MegaPay's actual spec
type megapayPayload struct {
	Reference   string  `json:"reference"`
	OrderID     int64   `json:"order_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	MpesaRef    string  `json:"mpesa_ref"`
	Signature   string  `json:"signature"`
}

func megapayHandler(store *db.Store, wiz *smmwiz.Client, secret string) http.HandlerFunc {
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

		// Verify HMAC signature if secret is configured
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
			// Not a successful payment — acknowledge but do nothing
			w.WriteHeader(http.StatusOK)
			return
		}

		ctx := r.Context()

		// Auto-confirm the transaction (system ID 0 = automated)
		if err := store.ConfirmTransaction(ctx, payload.OrderID, 0); err != nil {
			log.Printf("megapay confirmTransaction order %d: %v", payload.OrderID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Fulfill asynchronously
		go fulfillOrder(context.Background(), store, wiz, payload.OrderID)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}
}

func fulfillOrder(ctx context.Context, store *db.Store, wiz *smmwiz.Client, orderID int64) {
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
			store.UpdateOrderStatus(ctx, orderID, "failed", wizIDs)
			return
		}
		wizIDs = append(wizIDs, resp.Order)
	}

	store.UpdateOrderStatus(ctx, orderID, "processing", wizIDs)
	log.Printf("order %d fulfilled: wiz IDs %v", orderID, wizIDs)
}

func validSignature(body []byte, sig, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
