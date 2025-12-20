package models

import "time"

// Payment status constants
const (
	PaymentStatusDraft         = "draft"
	PaymentStatusScheduled     = "scheduled"
	PaymentStatusUpcoming      = "upcoming"
	PaymentStatusDue           = "due"
	PaymentStatusOverdue       = "overdue"
	PaymentStatusProcessing    = "processing"
	PaymentStatusPartiallyPaid = "partially_paid"
	PaymentStatusPaid          = "paid"
	PaymentStatusFailed        = "failed"
	PaymentStatusRefunded      = "refunded"
	PaymentStatusCancelled     = "cancelled"
	PaymentStatusWrittenOff    = "written_off"
)

// Payment recurrence constants
const (
	RecurrenceDaily      = "daily"
	RecurrenceWeekly     = "weekly"
	RecurrenceBiweekly   = "biweekly"
	RecurrenceMonthly    = "monthly"
	RecurrenceBimonthly  = "bimonthly"
	RecurrenceQuarterly  = "quarterly"
	RecurrenceSemiannual = "semiannual"
	RecurrenceAnnual     = "annual"
)

// Payment represents a payment extracted from an email
type Payment struct {
	ID                string
	AccountID         string
	Merchant          string
	Description       *string
	Amount            float64
	Currency          string
	Date              time.Time
	Recurrence        *string
	Status            string
	Category          *string
	ExternalReference *string
	Metadata          map[string]interface{}
	RawLlmResponse    map[string]interface{} // Store raw LLM response
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
