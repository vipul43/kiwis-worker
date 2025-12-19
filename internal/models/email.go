package models

import "time"

// Email represents raw email data for LLM fine-tuning
// This is a temporary table for collecting training data
type Email struct {
	ID             string
	AccountID      string
	GmailMessageID string
	GmailThreadID  *string
	From           *string
	To             *string
	CC             *string
	BCC            *string
	Subject        *string
	BodyText       *string
	BodyHTML       *string
	Snippet        *string
	ReceivedAt     *time.Time
	InternalDate   *time.Time
	Labels         []string
	RawHeaders     map[string]interface{} // JSON
	RawPayload     map[string]interface{} // JSON
	HasAttachments bool
	Attachments    []map[string]interface{} // JSON array
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
