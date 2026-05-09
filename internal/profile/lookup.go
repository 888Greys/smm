package profile

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Info holds public profile data scraped from og: meta tags.
type Info struct {
	Platform   string `json:"platform"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Bio        string `json:"bio"`
	Followers  string `json:"followers"`
	ProfileURL string `json:"profile_url"`
	AvatarURL  string `json:"avatar_url"`
	Found      bool   `json:"found"`
}

var httpClient = &http.Client{
	Timeout: 8 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

// ProfileURL returns the canonical URL for a username on a given platform.
func ProfileURL(platform, username string) string {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	switch platform {
	case "tiktok":
		return "https://www.tiktok.com/@" + username
	case "instagram":
		return "https://www.instagram.com/" + username + "/"
	case "youtube":
		return "https://www.youtube.com/@" + username
	default:
		return ""
	}
}

// Lookup fetches basic public profile info from og: meta tags.
// Returns Found=false on 404, blocked requests, or missing data — never errors on partial data.
func Lookup(platform, username string) (*Info, error) {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	profileURL := ProfileURL(platform, username)
	if profileURL == "" {
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return &Info{Platform: platform, Username: username, ProfileURL: profileURL, Found: false}, nil
	}
	// Mobile browser UA — better chance of getting server-rendered og: tags
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.216 Mobile Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return &Info{Platform: platform, Username: username, ProfileURL: profileURL, Found: false}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &Info{Platform: platform, Username: username, ProfileURL: profileURL, Found: false}, nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	html := string(body)

	name := extractOG(html, "title")
	desc := extractOG(html, "description")
	avatar := extractOG(html, "image")

	// Platform-specific name cleanup
	switch platform {
	case "youtube":
		name = strings.TrimSuffix(name, " - YouTube")
	case "instagram":
		if i := strings.Index(name, " •"); i > 0 {
			name = name[:i]
		}
		name = strings.TrimPrefix(name, "@")
	case "tiktok":
		if i := strings.Index(name, " | TikTok"); i > 0 {
			name = name[:i]
		}
		if i := strings.Index(name, " (@"); i > 0 {
			name = name[:i]
		}
	}

	followers := extractFollowers(platform, desc)
	found := name != "" && name != username && avatar != ""

	return &Info{
		Platform:   platform,
		Username:   username,
		Name:       name,
		Bio:        truncate(htmlUnescape(desc), 120),
		Followers:  followers,
		ProfileURL: profileURL,
		AvatarURL:  avatar,
		Found:      found,
	}, nil
}

// ── parsers ───────────────────────────────────────────────────────────────────

var (
	rePropFirst    = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:([^"']+)["'][^>]+content=["']([^"']*?)["']`)
	reContentFirst = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']*?)["'][^>]+property=["']og:([^"']+)["']`)
	reIGFollowers  = regexp.MustCompile(`(?i)([\d,\.]+[KkMmBb]?)\s*Followers`)
	reTTFollowers  = regexp.MustCompile(`(?i)([\d,\.]+[KkMmBb]?)\s*Followers`)
	reYTSubs       = regexp.MustCompile(`(?i)([\d,\.]+[KkMmBb]?)\s*[Ss]ubscribers?`)
)

func extractOG(html, prop string) string {
	for _, m := range rePropFirst.FindAllStringSubmatch(html, -1) {
		if strings.EqualFold(m[1], prop) {
			return htmlUnescape(m[2])
		}
	}
	for _, m := range reContentFirst.FindAllStringSubmatch(html, -1) {
		if strings.EqualFold(m[2], prop) {
			return htmlUnescape(m[1])
		}
	}
	return ""
}

func extractFollowers(platform, desc string) string {
	switch platform {
	case "instagram":
		if m := reIGFollowers.FindStringSubmatch(desc); len(m) > 1 {
			return m[1] + " followers"
		}
	case "tiktok":
		if m := reTTFollowers.FindStringSubmatch(desc); len(m) > 1 {
			return m[1] + " followers"
		}
	case "youtube":
		if m := reYTSubs.FindStringSubmatch(desc); len(m) > 1 {
			return m[1] + " subscribers"
		}
	}
	return ""
}

func htmlUnescape(s string) string {
	r := strings.NewReplacer(
		"&#39;", "'", "&amp;", "&", "&quot;", `"`,
		"&lt;", "<", "&gt;", ">", "&#x27;", "'", "&#x2F;", "/",
	)
	return r.Replace(s)
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n]) + "…"
}
