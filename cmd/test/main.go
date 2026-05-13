package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/aapom/smm/internal/smmpanel"
)

func main() {
	godotenv.Load()

	key := os.Getenv("MTP_API_KEY")
	if key == "" {
		log.Fatal("MTP_API_KEY not set in .env")
	}

	wiz := smmpanel.New(key)

	fmt.Println("=== BALANCE ===")
	bal, err := wiz.GetBalance()
	if err != nil {
		log.Fatalf("balance check failed: %v", err)
	}
	fmt.Printf("Available: %s %s\n\n", bal.Balance, bal.Currency)

	fmt.Println("=== SERVICES (filtered by platform) ===")
	services, err := wiz.GetServices()
	if err != nil {
		log.Fatalf("services fetch failed: %v", err)
	}

	// Filter groups: keyword → category slug
	groups := []struct {
		label   string
		keywords []string
	}{
		{"TIKTOK FOLLOWERS",   []string{"tiktok follower"}},
		{"TIKTOK VIEWS",       []string{"tiktok view", "tiktok video view"}},
		{"TIKTOK LIKES",       []string{"tiktok like"}},
		{"INSTAGRAM FOLLOWERS", []string{"instagram follower"}},
		{"INSTAGRAM LIKES",    []string{"instagram like"}},
		{"YOUTUBE SUBSCRIBERS", []string{"youtube subscriber", "youtube sub"}},
		{"YOUTUBE VIEWS",      []string{"youtube view"}},
	}

	type row struct {
		id      int
		name    string
		rate    string
		min     int
		max     int
		refill  bool
		cancel  bool
	}

	for _, g := range groups {
		var matches []row
		for _, s := range services {
			nameLower := strings.ToLower(s.Name)
			catLower := strings.ToLower(s.Category)
			hit := false
			for _, kw := range g.keywords {
				if strings.Contains(nameLower, kw) || strings.Contains(catLower, kw) {
					hit = true
					break
				}
			}
			if !hit {
				continue
			}
			matches = append(matches, row{
				id: s.Service, name: s.Name, rate: s.Rate,
				min: s.Min, max: s.Max, refill: s.Refill, cancel: s.Cancel,
			})
		}
		if len(matches) == 0 {
			continue
		}

		// Sort cheapest first
		sort.Slice(matches, func(i, j int) bool {
			ri, _ := strconv.ParseFloat(matches[i].rate, 64)
			rj, _ := strconv.ParseFloat(matches[j].rate, 64)
			return ri < rj
		})

		fmt.Printf("\n--- %s (%d services) ---\n", g.label, len(matches))
		fmt.Printf("%-7s %-60s %8s  %6s  %8s  %s\n", "ID", "Name", "Rate/1k", "Min", "Max", "Refill")
		fmt.Println(strings.Repeat("-", 100))
		for _, m := range matches {
			refillMark := ""
			if m.refill {
				refillMark = "✓"
			}
			name := m.name
			if len(name) > 58 {
				name = name[:55] + "..."
			}
			fmt.Printf("%-7d %-60s %8s  %6d  %8d  %s\n",
				m.id, name, "$"+m.rate, m.min, m.max, refillMark)
		}
	}

	fmt.Println("\n\n=== HOW TO PICK HIGH SUCCESS RATE SERVICES ===")
	fmt.Println(`
On morethanpanel.com → Best Sale Services tab:

  HIGH SUCCESS = high % AND high order count combined.
  The order count tells you the service has been battle-tested at scale.

  Signals to look for in the service NAME:
    ✓ "Lifetime Guaranteed" / "No Drop"  → stays delivered, no sudden drops
    ✓ "Refill"                            → panel will top up if count drops
    ✓ High order count (1000+)            → proven reliability
    ✓ Moderate speed (not "Instant")      → safer for account longevity

  Red flags:
    ✗ "Bot" / "Non-Drop" with no refill   → often drops
    ✗ Very cheap + very fast              → low quality, high drop rate
    ✗ Max is very low                     → limited capacity, may cancel

  Strategy: pick the 2nd or 3rd cheapest in each category that also has
  Refill=✓ and a name mentioning "HQ" or "Real" — those balance cost vs quality.
`)
}
