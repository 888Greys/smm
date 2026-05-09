package smmwiz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const apiURL = "https://smmwiz.com/api/v2"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- Request / Response types ---

type OrderRequest struct {
	Service      int
	Link         string
	Quantity     int
	Runs         int    // drip-feed: number of runs
	Interval     int    // drip-feed: minutes between runs
	Comments     string // custom comments, newline-separated
	Keywords     string // SEO keywords
	Username     string
	Usernames    string
	AnswerNumber string
	Groups       string
	// Subscription fields
	Min      int
	Max      int
	Posts    int
	OldPosts int
	Delay    int
	Expiry   string
}

type OrderResponse struct {
	Order int64  `json:"order"`
	Error string `json:"error"`
}

type StatusResponse struct {
	Charge     string `json:"charge"`
	StartCount string `json:"start_count"`
	Status     string `json:"status"`
	Remains    string `json:"remains"`
	Currency   string `json:"currency"`
	Error      string `json:"error"`
}

type Service struct {
	Service  int    `json:"service"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Rate     string `json:"rate"`
	Min      int    `json:"min"`
	Max      int    `json:"max"`
	Category string `json:"category"`
	Refill   bool   `json:"refill"`
	Cancel   bool   `json:"cancel"`
}

type BalanceResponse struct {
	Balance  string `json:"balance"`
	Currency string `json:"currency"`
	Error    string `json:"error"`
}

type RefillResponse struct {
	Refill int64  `json:"refill"`
	Error  string `json:"error"`
}

type RefillStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type CancelResponse struct {
	Cancel int64  `json:"cancel"`
	Error  string `json:"error"`
}

// --- Public methods ---

func (c *Client) AddOrder(req OrderRequest) (*OrderResponse, error) {
	params := c.baseParams("add")
	params.Set("service", strconv.Itoa(req.Service))
	params.Set("link", req.Link)

	if req.Quantity > 0 {
		params.Set("quantity", strconv.Itoa(req.Quantity))
	}
	if req.Runs > 0 {
		params.Set("runs", strconv.Itoa(req.Runs))
	}
	if req.Interval > 0 {
		params.Set("interval", strconv.Itoa(req.Interval))
	}
	if req.Comments != "" {
		params.Set("comments", req.Comments)
	}
	if req.Keywords != "" {
		params.Set("keywords", req.Keywords)
	}
	if req.Username != "" {
		params.Set("username", req.Username)
	}
	if req.Usernames != "" {
		params.Set("usernames", req.Usernames)
	}
	if req.AnswerNumber != "" {
		params.Set("answer_number", req.AnswerNumber)
	}
	if req.Groups != "" {
		params.Set("groups", req.Groups)
	}
	if req.Min > 0 {
		params.Set("min", strconv.Itoa(req.Min))
		params.Set("max", strconv.Itoa(req.Max))
		params.Set("posts", strconv.Itoa(req.Posts))
		params.Set("delay", strconv.Itoa(req.Delay))
	}
	if req.OldPosts > 0 {
		params.Set("old_posts", strconv.Itoa(req.OldPosts))
	}
	if req.Expiry != "" {
		params.Set("expiry", req.Expiry)
	}

	var resp OrderResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("smmwiz: %s", resp.Error)
	}
	return &resp, nil
}

func (c *Client) GetStatus(orderID int64) (*StatusResponse, error) {
	params := c.baseParams("status")
	params.Set("order", strconv.FormatInt(orderID, 10))

	var resp StatusResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("smmwiz: %s", resp.Error)
	}
	return &resp, nil
}

func (c *Client) GetMultiStatus(orderIDs []int64) (map[string]StatusResponse, error) {
	params := c.baseParams("status")
	params.Set("orders", joinInt64s(orderIDs))

	var resp map[string]StatusResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetServices() ([]Service, error) {
	params := c.baseParams("services")

	var resp []Service
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Refill(orderID int64) (*RefillResponse, error) {
	params := c.baseParams("refill")
	params.Set("order", strconv.FormatInt(orderID, 10))

	var resp RefillResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("smmwiz: %s", resp.Error)
	}
	return &resp, nil
}

func (c *Client) MultiRefill(orderIDs []int64) ([]RefillResponse, error) {
	params := c.baseParams("refill")
	params.Set("orders", joinInt64s(orderIDs))

	var resp []RefillResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetRefillStatus(refillID int64) (*RefillStatusResponse, error) {
	params := c.baseParams("refill_status")
	params.Set("refill", strconv.FormatInt(refillID, 10))

	var resp RefillStatusResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("smmwiz: %s", resp.Error)
	}
	return &resp, nil
}

func (c *Client) GetMultiRefillStatus(refillIDs []int64) (map[string]RefillStatusResponse, error) {
	params := c.baseParams("refill_status")
	params.Set("refills", joinInt64s(refillIDs))

	var resp map[string]RefillStatusResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CancelOrders(orderIDs []int64) ([]CancelResponse, error) {
	params := c.baseParams("cancel")
	params.Set("orders", joinInt64s(orderIDs))

	var resp []CancelResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetBalance() (*BalanceResponse, error) {
	params := c.baseParams("balance")

	var resp BalanceResponse
	if err := c.post(params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("smmwiz: %s", resp.Error)
	}
	return &resp, nil
}

// --- Internal helpers ---

func (c *Client) baseParams(action string) url.Values {
	p := url.Values{}
	p.Set("key", c.apiKey)
	p.Set("action", action)
	return p
}

func (c *Client) post(params url.Values, dest any) error {
	resp, err := c.httpClient.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("smmwiz http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("smmwiz read: %w", err)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("smmwiz parse: %w (body: %s)", err, string(body))
	}
	return nil
}

func joinInt64s(ids []int64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}
