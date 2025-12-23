package openrouter

import (
	"testing"
)

func TestCleanJSONResponse(t *testing.T) {
	client := NewClient("test-key")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `{"merchant": "Netflix"}`,
			expected: `{"merchant": "Netflix"}`,
		},
		{
			name:     "JSON with markdown code blocks",
			input:    "```json\n{\"merchant\": \"Netflix\"}\n```",
			expected: `{"merchant": "Netflix"}`,
		},
		{
			name:     "JSON with plain code blocks",
			input:    "```\n{\"merchant\": \"Netflix\"}\n```",
			expected: `{"merchant": "Netflix"}`,
		},
		{
			name:     "JSON with explanatory text before",
			input:    "Here is the payment information:\n{\"merchant\": \"Netflix\"}",
			expected: `{"merchant": "Netflix"}`,
		},
		{
			name:     "JSON with explanatory text after",
			input:    "{\"merchant\": \"Netflix\"}\nThis is a subscription payment.",
			expected: `{"merchant": "Netflix"}`,
		},
		{
			name:     "JSON with text before and after",
			input:    "No payment found. Output:\n{\"merchant\": null}\nEnd of response.",
			expected: `{"merchant": null}`,
		},
		{
			name:     "JSON with whitespace",
			input:    "  \n  {\"merchant\": \"Netflix\"}  \n  ",
			expected: `{"merchant": "Netflix"}`,
		},
		{
			name:     "null response",
			input:    "null",
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.cleanJSONResponse(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestIsValidPayment(t *testing.T) {
	client := NewClient("test-key")

	tests := []struct {
		name     string
		payment  PaymentData
		expected bool
	}{
		{
			name: "valid payment",
			payment: PaymentData{
				Merchant: "Netflix",
				Amount:   floatPtr(19.99),
				Currency: "USD",
				Date:     "2025-01-01T00:00:00",
				Status:   "upcoming",
			},
			expected: true,
		},
		{
			name: "missing merchant",
			payment: PaymentData{
				Merchant: "",
				Amount:   floatPtr(19.99),
				Currency: "USD",
				Date:     "2025-01-01T00:00:00",
				Status:   "upcoming",
			},
			expected: false,
		},
		{
			name: "nil amount",
			payment: PaymentData{
				Merchant: "Netflix",
				Amount:   nil,
				Currency: "USD",
				Date:     "2025-01-01T00:00:00",
				Status:   "upcoming",
			},
			expected: false,
		},
		{
			name: "zero amount",
			payment: PaymentData{
				Merchant: "Netflix",
				Amount:   floatPtr(0),
				Currency: "USD",
				Date:     "2025-01-01T00:00:00",
				Status:   "upcoming",
			},
			expected: false,
		},
		{
			name: "negative amount",
			payment: PaymentData{
				Merchant: "Netflix",
				Amount:   floatPtr(-10),
				Currency: "USD",
				Date:     "2025-01-01T00:00:00",
				Status:   "upcoming",
			},
			expected: false,
		},
		{
			name: "missing currency",
			payment: PaymentData{
				Merchant: "Netflix",
				Amount:   floatPtr(19.99),
				Currency: "",
				Date:     "2025-01-01T00:00:00",
				Status:   "upcoming",
			},
			expected: false,
		},
		{
			name: "missing date",
			payment: PaymentData{
				Merchant: "Netflix",
				Amount:   floatPtr(19.99),
				Currency: "USD",
				Date:     "",
				Status:   "upcoming",
			},
			expected: false,
		},
		{
			name: "missing status",
			payment: PaymentData{
				Merchant: "Netflix",
				Amount:   floatPtr(19.99),
				Currency: "USD",
				Date:     "2025-01-01T00:00:00",
				Status:   "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isValidPayment(tt.payment)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
