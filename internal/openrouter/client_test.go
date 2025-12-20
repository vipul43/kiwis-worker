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
			input:    `{"merchant_name": "Netflix"}`,
			expected: `{"merchant_name": "Netflix"}`,
		},
		{
			name:     "JSON with markdown code blocks",
			input:    "```json\n{\"merchant_name\": \"Netflix\"}\n```",
			expected: `{"merchant_name": "Netflix"}`,
		},
		{
			name:     "JSON with plain code blocks",
			input:    "```\n{\"merchant_name\": \"Netflix\"}\n```",
			expected: `{"merchant_name": "Netflix"}`,
		},
		{
			name:     "JSON with explanatory text before",
			input:    "Here is the payment information:\n{\"merchant_name\": \"Netflix\"}",
			expected: `{"merchant_name": "Netflix"}`,
		},
		{
			name:     "JSON with explanatory text after",
			input:    "{\"merchant_name\": \"Netflix\"}\nThis is a subscription payment.",
			expected: `{"merchant_name": "Netflix"}`,
		},
		{
			name:     "JSON with text before and after",
			input:    "No payment found. Output:\n{\"merchant_name\": null}\nEnd of response.",
			expected: `{"merchant_name": null}`,
		},
		{
			name:     "JSON with whitespace",
			input:    "  \n  {\"merchant_name\": \"Netflix\"}  \n  ",
			expected: `{"merchant_name": "Netflix"}`,
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
				MerchantName: "Netflix",
				Amount:       floatPtr(19.99),
				Currency:     "USD",
				Due:          "2025-01-01T00:00:00",
				Status:       "upcoming",
			},
			expected: true,
		},
		{
			name: "missing merchant name",
			payment: PaymentData{
				MerchantName: "",
				Amount:       floatPtr(19.99),
				Currency:     "USD",
				Due:          "2025-01-01T00:00:00",
				Status:       "upcoming",
			},
			expected: false,
		},
		{
			name: "nil amount",
			payment: PaymentData{
				MerchantName: "Netflix",
				Amount:       nil,
				Currency:     "USD",
				Due:          "2025-01-01T00:00:00",
				Status:       "upcoming",
			},
			expected: false,
		},
		{
			name: "zero amount",
			payment: PaymentData{
				MerchantName: "Netflix",
				Amount:       floatPtr(0),
				Currency:     "USD",
				Due:          "2025-01-01T00:00:00",
				Status:       "upcoming",
			},
			expected: false,
		},
		{
			name: "negative amount",
			payment: PaymentData{
				MerchantName: "Netflix",
				Amount:       floatPtr(-10),
				Currency:     "USD",
				Due:          "2025-01-01T00:00:00",
				Status:       "upcoming",
			},
			expected: false,
		},
		{
			name: "missing currency",
			payment: PaymentData{
				MerchantName: "Netflix",
				Amount:       floatPtr(19.99),
				Currency:     "",
				Due:          "2025-01-01T00:00:00",
				Status:       "upcoming",
			},
			expected: false,
		},
		{
			name: "missing due date",
			payment: PaymentData{
				MerchantName: "Netflix",
				Amount:       floatPtr(19.99),
				Currency:     "USD",
				Due:          "",
				Status:       "upcoming",
			},
			expected: false,
		},
		{
			name: "missing status",
			payment: PaymentData{
				MerchantName: "Netflix",
				Amount:       floatPtr(19.99),
				Currency:     "USD",
				Due:          "2025-01-01T00:00:00",
				Status:       "",
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
