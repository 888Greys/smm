package bot

import "github.com/aapom/smm/internal/models"

// Service IDs and wholesale rates from SMMWiz (USD per 1000 units):
// TikTok Followers ID:18612 — $1.08 — Real, 30D Refill
// TikTok Views     ID:19967 — $0.06 — No Drop, Instant
// TikTok Likes     ID:12350 — $0.09 — HQ Real
// IG Followers     ID:20888 — $0.55 — Real HQ, 30D Refill
// IG Likes         ID:19909 — $0.09 — Low Drop
// YT Subscribers   ID:9727  — $0.33 — Instant
// YT Views         ID:19646 — $0.34 — Lifetime Guaranteed
// Margins pre-calculated at 130 KES/USD

var Catalog = []models.Package{
	{
		ID: "test_ksh1", Name: "Test Package",
		Platform: models.PlatformTikTok, PriceKES: 1, MarginKES: 1,
		Description: "Test order — KES 1 only",
		Components:  []models.PackageComponent{{ServiceID: 19967, Quantity: 100}},
	},

	// ── Tier 1: Entry (KES 500) ──────────────────────────────────────────────
	{
		ID: "tiktok_flex", Name: "TikTok Flex",
		Platform: models.PlatformTikTok, PriceKES: 500, MarginKES: 414, Refillable: true,
		Description: "500 Followers + 2,000 Views",
		Components: []models.PackageComponent{
			{ServiceID: 18612, Quantity: 500},
			{ServiceID: 19967, Quantity: 2000},
		},
	},
	{
		ID: "ig_quick_start", Name: "IG Quick-Start",
		Platform: models.PlatformInstagram, PriceKES: 500, MarginKES: 463, Refillable: true,
		Description: "500 Followers + 100 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 20888, Quantity: 500},
			{ServiceID: 19909, Quantity: 100},
		},
	},

	// ── Tier 2: Growth ───────────────────────────────────────────────────────
	{
		ID: "tiktok_starter", Name: "TikTok Starter",
		Platform: models.PlatformTikTok, PriceKES: 1000, MarginKES: 791, Refillable: true,
		Description: "1,200 Followers + 5,000 Views + 100 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 18612, Quantity: 1200},
			{ServiceID: 19967, Quantity: 5000},
			{ServiceID: 12350, Quantity: 100},
		},
	},
	{
		ID: "ig_business_boost", Name: "IG Business Boost",
		Platform: models.PlatformInstagram, PriceKES: 800, MarginKES: 725, Refillable: true,
		Description: "1,000 Followers + 300 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 20888, Quantity: 1000},
			{ServiceID: 19909, Quantity: 300},
		},
	},

	// ── Tier 3: Combo Power ──────────────────────────────────────────────────
	{
		ID: "tiktok_viral_starter", Name: "TikTok Viral Starter",
		Platform: models.PlatformTikTok, PriceKES: 1500, MarginKES: 1178, Refillable: true,
		Description: "2,000 Followers + 5,000 Views + 200 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 18612, Quantity: 2000},
			{ServiceID: 19967, Quantity: 5000},
			{ServiceID: 12350, Quantity: 200},
		},
	},
	{
		ID: "ig_celebrity_pack", Name: "IG Celebrity Pack",
		Platform: models.PlatformInstagram, PriceKES: 2500, MarginKES: 2131, Refillable: true,
		Description: "5,000 Followers + 1,000 Likes (5-day drip delivery)",
		Components: []models.PackageComponent{
			{ServiceID: 20888, Quantity: 5000, Runs: 5, Interval: 1440}, // ~1000 followers/day over 5 days
			{ServiceID: 19909, Quantity: 1000},
		},
	},
	{
		ID: "youtube_kickstart", Name: "YouTube Kickstart",
		Platform: models.PlatformYouTube, PriceKES: 1500, MarginKES: 1443,
		Description: "300 Subscribers + 1,000 Views",
		Components: []models.PackageComponent{
			{ServiceID: 9727, Quantity: 300},
			{ServiceID: 19646, Quantity: 1000},
		},
	},
	{
		ID: "follower_booster", Name: "Follower Booster",
		Platform: models.PlatformInstagram, PriceKES: 600, MarginKES: 528, Refillable: true,
		Description: "1,000 IG Followers + 30-day Refill",
		Components: []models.PackageComponent{
			{ServiceID: 20888, Quantity: 1000},
		},
	},
}

// UpsellTarget returns the recommended upgrade package after an entry-tier purchase.
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

// RefillablePackageIDs returns the IDs of all packages that include a 30-day refill.
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
