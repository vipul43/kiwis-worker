package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/vipul43/kiwis-worker/internal/service"
)

type Client struct {
	clientID     string
	clientSecret string
}

func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// FetchMessageIDs fetches only message IDs from Gmail API (lightweight, fast)
func (c *Client) FetchMessageIDs(ctx context.Context, accessToken string, query string, maxResults int, pageToken string) (*service.MessageIDFetchResult, error) {
	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// Create Gmail service
	gmailService, err := gmail.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// List messages (only IDs, no full message fetch)
	listCall := gmailService.Users.Messages.List("me").Q(query).MaxResults(int64(maxResults))
	if pageToken != "" {
		listCall = listCall.PageToken(pageToken)
	}

	listResp, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	log.Printf("Gmail API returned %d message IDs (nextPageToken: %s)", len(listResp.Messages), listResp.NextPageToken)

	// Extract message IDs
	messageIDs := make([]string, 0, len(listResp.Messages))
	for _, msg := range listResp.Messages {
		messageIDs = append(messageIDs, msg.Id)
	}

	return &service.MessageIDFetchResult{
		MessageIDs:    messageIDs,
		NextPageToken: listResp.NextPageToken,
		TotalFetched:  len(messageIDs),
	}, nil
}

// FetchEmailByID fetches a single email by its Gmail message ID
func (c *Client) FetchEmailByID(ctx context.Context, accessToken string, messageID string) (*service.EmailMessage, error) {
	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// Create Gmail service
	gmailService, err := gmail.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Fetch full message by ID
	fullMsg, err := gmailService.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Parse message
	emailMsg, err := c.parseMessage(fullMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return &emailMsg, nil
}

// FetchEmails fetches emails from Gmail API
func (c *Client) FetchEmails(ctx context.Context, accessToken string, query string, maxResults int, pageToken string) (*service.EmailFetchResult, error) {
	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// Create Gmail service
	gmailService, err := gmail.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// List messages
	listCall := gmailService.Users.Messages.List("me").Q(query).MaxResults(int64(maxResults))
	if pageToken != "" {
		listCall = listCall.PageToken(pageToken)
	}

	listResp, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	log.Printf("Gmail API returned %d messages (nextPageToken: %s)", len(listResp.Messages), listResp.NextPageToken)

	// Fetch full message details for each message
	messages := make([]service.EmailMessage, 0, len(listResp.Messages))
	for _, msg := range listResp.Messages {
		fullMsg, err := gmailService.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("Warning: failed to get message %s: %v", msg.Id, err)
			continue
		}

		emailMsg, err := c.parseMessage(fullMsg)
		if err != nil {
			log.Printf("Warning: failed to parse message %s: %v", msg.Id, err)
			continue
		}

		messages = append(messages, emailMsg)
	}

	return &service.EmailFetchResult{
		Messages:      messages,
		NextPageToken: listResp.NextPageToken,
		TotalFetched:  len(messages),
	}, nil
}

// parseMessage parses Gmail message into EmailMessage struct with all fields
func (c *Client) parseMessage(msg *gmail.Message) (service.EmailMessage, error) {
	emailMsg := service.EmailMessage{
		ID:             msg.Id,
		ThreadID:       msg.ThreadId,
		Snippet:        msg.Snippet,
		Labels:         msg.LabelIds,
		RawHeaders:     make(map[string]interface{}),
		RawPayload:     make(map[string]interface{}),
		HasAttachments: false,
		Attachments:    []map[string]interface{}{},
	}

	// Parse internal date (milliseconds since epoch)
	if msg.InternalDate > 0 {
		emailMsg.InternalDate = time.UnixMilli(msg.InternalDate)
	}

	// Parse all headers and store in RawHeaders
	for _, header := range msg.Payload.Headers {
		emailMsg.RawHeaders[header.Name] = header.Value

		switch header.Name {
		case "Subject":
			emailMsg.Subject = header.Value
		case "From":
			emailMsg.From = header.Value
		case "To":
			emailMsg.To = header.Value
		case "Cc":
			emailMsg.CC = header.Value
		case "Bcc":
			emailMsg.BCC = header.Value
		case "Date":
			parsedDate, err := parseEmailDate(header.Value)
			if err != nil {
				log.Printf("Warning: failed to parse date '%s': %v", header.Value, err)
			} else {
				emailMsg.Date = parsedDate
			}
		}
	}

	// Extract body (text and HTML)
	bodyText, bodyHTML := c.extractBodies(msg.Payload)
	emailMsg.BodyText = bodyText
	emailMsg.BodyHTML = bodyHTML

	// Extract attachments info
	attachments := c.extractAttachments(msg.Payload)
	if len(attachments) > 0 {
		emailMsg.HasAttachments = true
		emailMsg.Attachments = attachments
	}

	// Store raw payload structure (without actual attachment data)
	emailMsg.RawPayload = map[string]interface{}{
		"mimeType": msg.Payload.MimeType,
		"filename": msg.Payload.Filename,
		"partId":   msg.Payload.PartId,
	}

	return emailMsg, nil
}

// extractBodies extracts both text and HTML bodies from message payload
func (c *Client) extractBodies(payload *gmail.MessagePart) (string, string) {
	var textPlain, textHTML string

	// Check if body is in the main payload
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			switch payload.MimeType {
			case "text/plain":
				textPlain = string(decoded)
			case "text/html":
				textHTML = string(decoded)
			}
		}
	}

	// Recursively extract from parts
	c.extractBodiesFromParts(payload.Parts, &textPlain, &textHTML)

	return textPlain, textHTML
}

// extractBodiesFromParts recursively extracts text and HTML from message parts
func (c *Client) extractBodiesFromParts(parts []*gmail.MessagePart, textPlain, textHTML *string) {
	for _, part := range parts {
		if part.Body != nil && part.Body.Data != "" {
			decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err == nil {
				if part.MimeType == "text/plain" && *textPlain == "" {
					*textPlain = string(decoded)
				} else if part.MimeType == "text/html" && *textHTML == "" {
					*textHTML = string(decoded)
				}
			}
		}

		// Recursively check nested parts
		if len(part.Parts) > 0 {
			c.extractBodiesFromParts(part.Parts, textPlain, textHTML)
		}
	}
}

// extractAttachments extracts attachment metadata from message payload
func (c *Client) extractAttachments(payload *gmail.MessagePart) []map[string]interface{} {
	attachments := []map[string]interface{}{}
	c.extractAttachmentsFromParts(payload.Parts, &attachments)
	return attachments
}

// extractAttachmentsFromParts recursively extracts attachment info from parts
func (c *Client) extractAttachmentsFromParts(parts []*gmail.MessagePart, attachments *[]map[string]interface{}) {
	for _, part := range parts {
		// Check if this part is an attachment
		if part.Filename != "" && part.Body != nil {
			attachment := map[string]interface{}{
				"filename": part.Filename,
				"mimeType": part.MimeType,
				"size":     part.Body.Size,
			}
			if part.Body.AttachmentId != "" {
				attachment["attachmentId"] = part.Body.AttachmentId
			}
			*attachments = append(*attachments, attachment)
		}

		// Recursively check nested parts
		if len(part.Parts) > 0 {
			c.extractAttachmentsFromParts(part.Parts, attachments)
		}
	}
}

// RefreshAccessToken refreshes the OAuth2 access token
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (*service.TokenRefreshResult, error) {
	config := &oauth2.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Refresh the token
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	result := &service.TokenRefreshResult{
		AccessToken: newToken.AccessToken,
		ExpiresAt:   newToken.Expiry,
	}

	// Check if refresh token was rotated
	if newToken.RefreshToken != "" && newToken.RefreshToken != refreshToken {
		result.RefreshToken = newToken.RefreshToken
	} else {
		result.RefreshToken = refreshToken // Keep the same refresh token
	}

	log.Printf("Token refreshed successfully, expires at: %s", result.ExpiresAt)

	return result, nil
}

// parseEmailDate parses various email date formats
func parseEmailDate(dateStr string) (time.Time, error) {
	// Common email date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2 Jan 2006 15:04:05 -0700",
		time.RFC3339,
	}

	// Clean up the date string
	dateStr = strings.TrimSpace(dateStr)

	// Remove timezone name in parentheses (e.g., "(UTC)", "(PST)")
	// Gmail sometimes adds this after the numeric offset
	if idx := strings.Index(dateStr, " ("); idx != -1 {
		dateStr = dateStr[:idx]
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
