package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/aapom/smm/internal/bot"
	"github.com/aapom/smm/internal/db"
	"github.com/aapom/smm/internal/megapay"
	"github.com/aapom/smm/internal/smmwiz"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := db.NewStore(ctx, mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer store.Close()

	wiz := smmwiz.New(mustEnv("SMMWIZ_API_KEY"))
	pay := megapay.New(mustEnv("MEGAPAY_API_KEY"), mustEnv("MEGAPAY_EMAIL"))
	adminIDs := parseAdminIDs(mustEnv("ADMIN_TELEGRAM_IDS"))

	var proofChannelID int64
	if ch := os.Getenv("SOCIAL_PROOF_CHANNEL_ID"); ch != "" {
		proofChannelID, _ = strconv.ParseInt(ch, 10, 64)
	}

	// Optional admin notifier bot — set both vars to enable traffic tracking
	var notifier *bot.AdminNotifier
	adminBotToken := os.Getenv("ADMIN_BOT_TOKEN")
	adminChatIDStr := os.Getenv("ADMIN_CHAT_ID")
	if adminBotToken != "" && adminChatIDStr != "" {
		adminChatID, err := strconv.ParseInt(adminChatIDStr, 10, 64)
		if err != nil {
			log.Fatalf("ADMIN_CHAT_ID invalid: %v", err)
		}
		notifier, err = bot.NewAdminNotifier(adminBotToken, adminChatID)
		if err != nil {
			log.Fatalf("admin notifier: %v", err)
		}
	} else {
		log.Println("admin notifier disabled (ADMIN_BOT_TOKEN / ADMIN_CHAT_ID not set)")
	}

	b, err := bot.New(mustEnv("TELEGRAM_BOT_TOKEN"), wiz, pay, store, adminIDs, proofChannelID, notifier)
	if err != nil {
		log.Fatalf("bot init: %v", err)
	}

	b.Run(ctx)
	log.Println("bot stopped")
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
