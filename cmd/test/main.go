package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/aapom/smm/internal/smmwiz"
)

func main() {
	godotenv.Load()

	key := os.Getenv("SMMWIZ_API_KEY")
	if key == "" {
		log.Fatal("SMMWIZ_API_KEY not set in .env")
	}

	wiz := smmwiz.New(key)

	// 1. Check balance
	fmt.Println("=== BALANCE ===")
	bal, err := wiz.GetBalance()
	if err != nil {
		log.Fatalf("balance check failed: %v", err)
	}
	fmt.Printf("Available: %s %s\n\n", bal.Balance, bal.Currency)

	// 2. List TikTok services
	fmt.Println("=== TIKTOK SERVICES ===")
	services, err := wiz.GetServices()
	if err != nil {
		log.Fatalf("services fetch failed: %v", err)
	}

	for _, s := range services {
		if strings.Contains(strings.ToLower(s.Name), "tiktok") ||
			strings.Contains(strings.ToLower(s.Category), "tiktok") {
			fmt.Printf("ID: %-6d | %-50s | Rate: %s | Min: %s Max: %s | Refill: %v\n",
				s.Service, s.Name, s.Rate, s.Min, s.Max, s.Refill)
		}
	}
}
