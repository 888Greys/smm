package megapay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	stkURL    = "https://megapay.co.ke/backend/v1/initiatestk"
	statusURL = "https://megapay.co.ke/backend/v1/transactionstatus"
)

type Client struct {
	apiKey     string
	email      string
	httpClient *http.Client
}

func New(apiKey, email string) *Client {
	return &Client{
		apiKey: apiKey,
		email:  email,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type STKRequest struct {
	APIKey    string `json:"api_key"`
	Email     string `json:"email"`
	Amount    string `json:"amount"`
	Msisdn    string `json:"msisdn"`
	Reference string `json:"reference"`
}

type STKResponse struct {
	Success              string `json:"success"`
	Message              string `json:"message"`
	TransactionRequestID string `json:"transaction_request_id"`
	ResponseCode         string `json:"ResponseCode"`
}

type StatusRequest struct {
	APIKey               string `json:"api_key"`
	Email                string `json:"email"`
	TransactionRequestID string `json:"transaction_request_id"`
}

type StatusResponse struct {
	ResultCode        string `json:"ResultCode"`
	ResultDesc        string `json:"ResultDesc"`
	TransactionStatus string `json:"TransactionStatus"`
	TransactionCode   string `json:"TransactionCode"`
	TransactionReceipt string `json:"TransactionReceipt"`
	TransactionAmount string `json:"TransactionAmount"`
	Msisdn            string `json:"Msisdn"`
}

// InitiateSTK sends an M-Pesa STK push to the client's phone
func (c *Client) InitiateSTK(amountKES int, phone, reference string) (*STKResponse, error) {
	req := STKRequest{
		APIKey:    c.apiKey,
		Email:     c.email,
		Amount:    fmt.Sprintf("%d", amountKES),
		Msisdn:    phone,
		Reference: reference,
	}

	var resp STKResponse
	if err := c.post(stkURL, req, &resp); err != nil {
		return nil, err
	}
	if resp.ResponseCode != "0" && resp.Success != "200" {
		return nil, fmt.Errorf("megapay STK: %s", resp.Message)
	}
	return &resp, nil
}

// CheckStatus polls the status of a transaction
func (c *Client) CheckStatus(transactionRequestID string) (*StatusResponse, error) {
	req := StatusRequest{
		APIKey:               c.apiKey,
		Email:                c.email,
		TransactionRequestID: transactionRequestID,
	}

	var resp StatusResponse
	if err := c.postDebug(statusURL, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) post(url string, body any, dest any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("megapay http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("megapay read: %w", err)
	}

	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("megapay parse: %w (body: %s)", err, string(raw))
	}
	return nil
}

// postDebug is like post but logs the raw response body for debugging
func (c *Client) postDebug(url string, body any, dest any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("megapay http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("megapay read: %w", err)
	}

	log.Printf("megapay raw response: %s", string(raw))

	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("megapay parse: %w (body: %s)", err, string(raw))
	}
	return nil
}
