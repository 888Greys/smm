package bot

import "github.com/aapom/smm/internal/models"

// Catalog is the retail package menu shown to clients.
// ServiceID values must match actual SMMWiz service IDs from your panel.
// Replace the placeholder IDs after running /services to list them.
var Catalog = []models.Package{
	{
		ID:          "tiktok_viral_starter",
		Name:        "TikTok Viral Starter",
		Platform:    models.PlatformTikTok,
		PriceKES:    1500,
		Description: "2,000 Followers + 5,000 Views + 200 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 0, Quantity: 2000}, // TikTok Followers — set real ID
			{ServiceID: 0, Quantity: 5000}, // TikTok Views    — set real ID
			{ServiceID: 0, Quantity: 200},  // TikTok Likes    — set real ID
		},
	},
	{
		ID:          "ig_business_boost",
		Name:        "IG Business Boost",
		Platform:    models.PlatformInstagram,
		PriceKES:    1500,
		Description: "1,500 HQ Followers + 300 Likes",
		Components: []models.PackageComponent{
			{ServiceID: 0, Quantity: 1500}, // IG Followers — set real ID
			{ServiceID: 0, Quantity: 300},  // IG Likes     — set real ID
		},
	},
	{
		ID:          "youtube_kickstart",
		Name:        "YouTube Kickstart",
		Platform:    models.PlatformYouTube,
		PriceKES:    1500,
		Description: "300 Subscribers + 1,000 Watch Views",
		Components: []models.PackageComponent{
			{ServiceID: 0, Quantity: 300},  // YT Subscribers — set real ID
			{ServiceID: 0, Quantity: 1000}, // YT Views       — set real ID
		},
	},
	{
		ID:          "follower_booster",
		Name:        "Follower Booster",
		Platform:    models.PlatformInstagram,
		PriceKES:    600,
		Description: "1,000 Followers + 30-day Refill",
		Components: []models.PackageComponent{
			{ServiceID: 0, Quantity: 1000}, // IG Followers with refill — set real ID
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
