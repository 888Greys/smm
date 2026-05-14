package bot

import "github.com/aapom/smm/internal/models"

// Service IDs and wholesale rates from morethanpanel.com (USD/1000 at 130 KES/USD):
// TikTok Followers ID:5760  — $2.44 — 30 Day Refill, 5-10K/Day
// TikTok Views     ID:9121  — $0.04 — 30 Day Refill, 10-100K/Day
// TikTok Likes     ID:2699  — $0.32 — 30 Day Refill, 5-50K/Day
// IG Followers     ID:5440  — $0.35 — 30 Day Refill, 10-50K/Day
// IG Likes         ID:2916  — $0.10 — 30 Day Refill, 5-10K/Day
// YT Subscribers   ID:7494  — $0.70 — No Refill, 10-50K/Day
// YT Views         ID:6003  — $0.41 — Lifetime Guaranteed

var Catalog = []models.Package{
	{
		ID: "test_ksh1", Name: "Test Package",
		Platform: models.PlatformTikTok, Category: "tiktok",
		PriceKES: 1, MarginKES: 1,
		Description: "Test order — KES 1 only",
		Components:  []models.PackageComponent{{ServiceID: 9121, Quantity: 100}},
	},

	// ── TikTok ───────────────────────────────────────────────────────────────
	{
		ID: "tiktok_flex", Name: "TikTok Quick-Start",
		Platform: models.PlatformTikTok, Category: "tiktok",
		PriceKES: 1, MarginKES: 1, Refillable: true,
		Description: "500 Real Followers + 2,000 Views",
		Components: []models.PackageComponent{
			{ServiceID: 5760, Quantity: 500},
			{ServiceID: 9121, Quantity: 2000},
		},
	},
	{
		ID: "tiktok_starter", Name: "TikTok Starter",
		Platform: models.PlatformTikTok, Category: "tiktok",
		PriceKES: 1000, MarginKES: 589, Refillable: true,
		Description: "1,200 Followers + 5,000 Views + 100 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 5760, Quantity: 1200},
			{ServiceID: 9121, Quantity: 5000},
			{ServiceID: 2699, Quantity: 100},
		},
	},
	{
		ID: "tiktok_viral_starter", Name: "TikTok Viral Starter",
		Platform: models.PlatformTikTok, Category: "tiktok",
		PriceKES: 1500, MarginKES: 831, Refillable: true,
		Description: "2,000 Followers + 5,000 Views + 200 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 5760, Quantity: 2000},
			{ServiceID: 9121, Quantity: 5000},
			{ServiceID: 2699, Quantity: 200},
		},
	},

	// ── Instagram ─────────────────────────────────────────────────────────────
	{
		ID: "ig_quick_start", Name: "IG Quick-Start",
		Platform: models.PlatformInstagram, Category: "instagram",
		PriceKES: 500, MarginKES: 476, Refillable: true,
		Description: "500 Real Followers + 100 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 5440, Quantity: 500},
			{ServiceID: 2916, Quantity: 100},
		},
	},
	{
		ID: "ig_business_boost", Name: "IG Business Boost",
		Platform: models.PlatformInstagram, Category: "instagram",
		PriceKES: 800, MarginKES: 751, Refillable: true,
		Description: "1,000 Followers + 300 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 5440, Quantity: 1000},
			{ServiceID: 2916, Quantity: 300},
		},
	},
	{
		ID: "follower_booster", Name: "Follower Booster",
		Platform: models.PlatformInstagram, Category: "instagram",
		PriceKES: 600, MarginKES: 555, Refillable: true,
		Description: "1,000 Followers + 30-day Refill Guarantee",
		Components: []models.PackageComponent{
			{ServiceID: 5440, Quantity: 1000},
		},
	},
	{
		ID: "ig_celebrity_pack", Name: "IG Celebrity Pack",
		Platform: models.PlatformInstagram, Category: "instagram",
		PriceKES: 2500, MarginKES: 2260, Refillable: true,
		Description: "5,000 Followers + 1,000 Likes (5-day drip delivery)",
		Components: []models.PackageComponent{
			{ServiceID: 5440, Quantity: 5000, Runs: 5, Interval: 1440},
			{ServiceID: 2916, Quantity: 1000},
		},
	},

	// ── YouTube ───────────────────────────────────────────────────────────────
	{
		ID: "youtube_kickstart", Name: "YouTube Kickstart",
		Platform: models.PlatformYouTube, Category: "youtube",
		PriceKES: 1500, MarginKES: 1419,
		Description: "300 Subscribers + 1,000 Views",
		Components: []models.PackageComponent{
			{ServiceID: 7494, Quantity: 300},
			{ServiceID: 6003, Quantity: 1000},
		},
	},

	// ── Combo Deals ───────────────────────────────────────────────────────────
	{
		ID: "viral_creator_combo", Name: "Viral Creator Combo",
		Platform: models.PlatformTikTok, Category: "combo",
		PriceKES: 2500, MarginKES: 1476, Refillable: true,
		Description: "3,000 TikTok Followers + 10,000 Views + 500 Likes (drip-fed for safety)",
		Components: []models.PackageComponent{
			{ServiceID: 5760, Quantity: 3000, Runs: 6, Interval: 1440}, // ~500/day
			{ServiceID: 9121, Quantity: 10000},
			{ServiceID: 2699, Quantity: 500},
		},
	},
}

// CategoryPackages returns packages filtered by category slug.
func CategoryPackages(category string) []models.Package {
	var result []models.Package
	for _, p := range Catalog {
		if p.ID == "test_ksh1" {
			continue
		}
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

// UpsellTarget returns the recommended upgrade after an entry-tier purchase.
func UpsellTarget(packageID string) (models.Package, bool) {
	targets := map[string]string{
		"tiktok_flex":    "tiktok_viral_starter",
		"ig_quick_start": "ig_business_boost",
	}
	targetID, ok := targets[packageID]
	if !ok {
		return models.Package{}, false
	}
	return GetPackage(targetID)
}

// RefillablePackageIDs returns IDs of all packages with 30-day refill.
func RefillablePackageIDs() []string {
	var ids []string
	for _, p := range Catalog {
		if p.Refillable {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func GetPackage(id string) (models.Package, bool) {
	for _, p := range Catalog {
		if p.ID == id {
			return p, true
		}
	}
	return models.Package{}, false
}
