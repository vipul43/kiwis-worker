package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	OpenRouterAPIURL = "https://openrouter.ai/api/v1/chat/completions"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      *string // Optional: if nil, uses OpenRouter account default
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes timeout for LLM calls (free models are slow)
		},
		model: nil, // Use OpenRouter account default
	}
}

// SetModel sets a specific model to use (optional)
func (c *Client) SetModel(model string) {
	c.model = &model
}

// EmailData represents the email data to extract payment from
type EmailData struct {
	From    string
	Subject string
	Body    string
}

// PaymentData represents the extracted payment information
type PaymentData struct {
	MerchantName      string                 `json:"merchant_name"`
	Description       string                 `json:"description"`
	Amount            *float64               `json:"amount"`
	Currency          string                 `json:"currency"`
	Due               string                 `json:"due"`
	Recurrence        *string                `json:"recurrence"`
	Status            string                 `json:"status"`
	Category          string                 `json:"category"`
	ExternalReference string                 `json:"external_reference"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// BatchExtractPayments extracts payment information from multiple emails using OpenRouter batch API
func (c *Client) BatchExtractPayments(ctx context.Context, emails []EmailData) ([]PaymentData, []map[string]interface{}, error) {
	if len(emails) == 0 {
		return nil, nil, nil
	}

	// For now, process sequentially (OpenRouter free tier may not support true batching)
	// TODO: Implement true batch API when available
	results := make([]PaymentData, 0, len(emails))
	rawResponses := make([]map[string]interface{}, 0, len(emails))

	for _, email := range emails {
		payment, rawResp, err := c.ExtractPayment(ctx, email)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract payment: %w", err)
		}

		// Only add if it's a valid payment (has required fields)
		if payment != nil {
			results = append(results, *payment)
			rawResponses = append(rawResponses, rawResp)
		} else {
			// Not a payment email, add nil placeholder
			results = append(results, PaymentData{})
			rawResponses = append(rawResponses, rawResp)
		}
	}

	return results, rawResponses, nil
}

// ExtractPayment extracts payment information from a single email
func (c *Client) ExtractPayment(ctx context.Context, email EmailData) (*PaymentData, map[string]interface{}, error) {
	prompt := c.buildPrompt(email)

	reqBody := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// Only include model if explicitly set, otherwise use OpenRouter account default
	if c.model != nil {
		reqBody["model"] = *c.model
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", OpenRouterAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenRouter response
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, nil, fmt.Errorf("no response from LLM")
	}

	content := apiResp.Choices[0].Message.Content

	// Store raw response for audit
	var rawResponse map[string]interface{}
	_ = json.Unmarshal(body, &rawResponse)

	// Clean the content (remove markdown code blocks if present)
	cleanedContent := c.cleanJSONResponse(content)

	// Parse payment data from LLM response
	var paymentData PaymentData
	if err := json.Unmarshal([]byte(cleanedContent), &paymentData); err != nil {
		return nil, rawResponse, fmt.Errorf("failed to parse payment JSON: %w", err)
	}

	// Validate required fields
	if !c.isValidPayment(paymentData) {
		// Not a payment email or missing required fields
		return nil, rawResponse, nil
	}

	return &paymentData, rawResponse, nil
}

// cleanJSONResponse removes markdown code blocks and extra whitespace from LLM response
func (c *Client) cleanJSONResponse(content string) string {
	content = strings.TrimSpace(content)

	// Find the first { and last } to extract just the JSON object
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")

	if startIdx == -1 || endIdx == -1 || startIdx > endIdx {
		// No valid JSON found, return as is and let JSON parser fail with proper error
		return content
	}

	// Extract just the JSON object
	jsonContent := content[startIdx : endIdx+1]

	return strings.TrimSpace(jsonContent)
}

// buildPrompt builds the LLM prompt from email data
func (c *Client) buildPrompt(email EmailData) string {
	return fmt.Sprintf(`You are an AI that extracts structured upcoming-payment information from emails, messages, invoices, or notifications.

Your job is to analyze the given text and return a STRICT JSON object containing the fields required to populate the upcoming_payments table.

### OUTPUT FORMAT (STRICT JSON ONLY)
Return JSON with these keys:

{
  "merchant_name": "",
  "description": "",
  "amount": null,
  "currency": "",
  "due": "",
  "recurrence": null,
  "status": "",
  "category": "",
  "external_reference": "",
  "metadata": {}
}

### FIELD DEFINITIONS

merchant_name  
- The business or entity requesting payment (e.g., "Netflix", "Amazon Pay", "HDFC Bank").

description  
- Short natural-language description of what the payment is for.

amount  
- Numeric value only. Do NOT include commas or currency symbols.

currency  
- Infer from text: INR, USD, EUR, GBP, etc. Default to INR if unclear.

due  
- The next due date/time in ISO 8601 format: YYYY-MM-DDTHH:MM:SS  
  If only a date is available, use "T00:00:00".

recurrence  
- one of: null, "monthly", "yearly", "weekly", "daily", "quarterly", "semiannual"
- If subscription-like, infer the correct recurrence.

status  
- one of: "upcoming", "due_soon", "overdue", "paid", "cancelled"  
- Default: "upcoming"

category  
- One of: "subscription", "utility", "emi", "credit_card_bill", "loan", "insurance", "rent", "misc"  
- Infer logically.

external_reference  
- Invoice number, subscription ID, bill number, reference ID, order number, UTR, etc.  
- Null if unavailable.

metadata  
- JSON object with ANY additional important details:
  - billing period
  - statement date
  - last payment date
  - plan name
  - card used
  - UTR / transaction hash
  - customer ID  
  - etc.

### CRITICAL RULES
- Output ONLY the JSON object, no explanations.
- All values must exist. Use null if missing.
- Never hallucinate merchant names; infer only from text.
- If multiple amounts appear, pick the one associated with the upcoming payment.
- If due date not found, set "due": null.

### Now extract the payment JSON from this input:

From: %s
Subject: %s

%s`, email.From, email.Subject, email.Body)
}

// isValidPayment checks if the payment data has all required fields
func (c *Client) isValidPayment(payment PaymentData) bool {
	// Required fields: merchant_name, amount, currency, date, status
	if payment.MerchantName == "" {
		return false
	}
	if payment.Amount == nil || *payment.Amount <= 0 {
		return false
	}
	if payment.Currency == "" {
		return false
	}
	if payment.Due == "" {
		return false
	}
	if payment.Status == "" {
		return false
	}
	return true
}
