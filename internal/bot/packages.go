package bot

import "github.com/aapom/smm/internal/models"

// Rates per 1000 units (USD) from SMMWiz services list:
//
// TikTok Followers ID:18612 — $1.08/1000 — Real, 30 Days Refill, 5-10K/D
// TikTok Views     ID:19967 — $0.06/1000 — No Drop, Instant
// TikTok Likes     ID:12350 — $0.09/1000 — HQ Real, 50K/D
// IG Followers     ID:20888 — $0.55/1000 — Real HQ, 30 Days Refill, 50K/D
// IG Likes         ID:19909 — $0.09/1000 — Low Drop, 100K+/D
// YT Subscribers   ID:9727  — $0.33/1000 — No Refill, 0-6HRS
// YT Views         ID:19646 — $0.34/1000 — Lifetime Guaranteed, Instant

var Catalog = []models.Package{
	{
		ID:          "tiktok_viral_starter",
		Name:        "TikTok Viral Starter",
		Platform:    models.PlatformTikTok,
		PriceKES:    1500,
		Description: "2,000 Followers + 5,000 Views + 200 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 18612, Quantity: 2000}, // TikTok Followers — Real, 30D Refill
			{ServiceID: 19967, Quantity: 5000}, // TikTok Views — No Drop, Instant
			{ServiceID: 12350, Quantity: 200},  // TikTok Likes — HQ Real
		},
	},
	{
		ID:          "ig_business_boost",
		Name:        "IG Business Boost",
		Platform:    models.PlatformInstagram,
		PriceKES:    1500,
		Description: "1,500 HQ Followers + 300 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 20888, Quantity: 1500}, // IG Followers — Real HQ, 30D Refill
			{ServiceID: 19909, Quantity: 300},  // IG Likes — Low Drop
		},
	},
	{
		ID:          "youtube_kickstart",
		Name:        "YouTube Kickstart",
		Platform:    models.PlatformYouTube,
		PriceKES:    1500,
		Description: "300 Subscribers + 1,000 Views",
		Components: []models.PackageComponent{
			{ServiceID: 9727,  Quantity: 300},  // YT Subscribers — Instant
			{ServiceID: 19646, Quantity: 1000}, // YT Views — Lifetime Guaranteed
		},
	},
	{
		ID:          "follower_booster",
		Name:        "Follower Booster",
		Platform:    models.PlatformInstagram,
		PriceKES:    600,
		Description: "1,000 Followers + 30-day Refill",
		Components: []models.PackageComponent{
			{ServiceID: 20888, Quantity: 1000}, // IG Followers — Real HQ, 30D Refill
		},
	},
}

func GetPackage(id string) (models.Package, bool) {
	for _, p := range Catalog {
		if p.ID == id {
			return p, true
		}
	}
	return models.Package{}, false
}
