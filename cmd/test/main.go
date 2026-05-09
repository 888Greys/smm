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

	fmt.Println("=== BALANCE ===")
	bal, err := wiz.GetBalance()
	if err != nil {
		log.Fatalf("balance check failed: %v", err)
	}
	fmt.Printf("Available: %s %s\n\n", bal.Balance, bal.Currency)

	fmt.Println("=== SERVICES ===")
	services, err := wiz.GetServices()
	if err != nil {
		log.Fatalf("services fetch failed: %v", err)
	}

	keywords := []string{"tiktok", "instagram follower", "youtube sub", "youtube view"}
	for _, kw := range keywords {
		fmt.Printf("\n--- %s ---\n", strings.ToUpper(kw))
		for _, s := range services {
			name := strings.ToLower(s.Name)
			cat := strings.ToLower(s.Category)
			if strings.Contains(name, kw) || strings.Contains(cat, kw) {
				fmt.Printf("ID:%-6d Rate:$%-8s Min:%-8d Max:%-10d %s\n",
					s.Service, s.Rate, s.Min, s.Max, s.Name)
			}
		}
	}
}
