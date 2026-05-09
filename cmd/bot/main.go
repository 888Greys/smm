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
	"github.com/aapom/smm/internal/smmwiz"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	tgToken := mustEnv("TELEGRAM_BOT_TOKEN")
	wizKey := mustEnv("SMMWIZ_API_KEY")
	adminStr := mustEnv("ADMIN_TELEGRAM_IDS") // comma-separated

	adminIDs := parseAdminIDs(adminStr)

	wiz := smmwiz.New(wizKey)

	// TODO: wire up real DB store (see internal/db)
	// For now pass nil — replace with db.NewStore(connString)
	b, err := bot.New(tgToken, wiz, nil, adminIDs)
	if err != nil {
		log.Fatalf("bot init: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
