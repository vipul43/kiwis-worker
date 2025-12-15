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

// parseMessage parses Gmail message into EmailMessage struct
func (c *Client) parseMessage(msg *gmail.Message) (service.EmailMessage, error) {
	emailMsg := service.EmailMessage{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
	}

	// Parse headers
	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "Subject":
			emailMsg.Subject = header.Value
		case "From":
			emailMsg.From = header.Value
		case "Date":
			// Parse date
			parsedDate, err := parseEmailDate(header.Value)
			if err != nil {
				log.Printf("Warning: failed to parse date '%s': %v", header.Value, err)
				emailMsg.Date = time.Now() // Fallback to now
			} else {
				emailMsg.Date = parsedDate
			}
		}
	}

	// Extract body
	body, err := c.extractBody(msg.Payload)
	if err != nil {
		log.Printf("Warning: failed to extract body for message %s: %v", msg.Id, err)
	}
	emailMsg.Body = body

	return emailMsg, nil
}

// extractBody extracts email body from message payload
func (c *Client) extractBody(payload *gmail.MessagePart) (string, error) {
	// Check if body is in the main payload
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode body: %w", err)
		}
		return string(decoded), nil
	}

	// Check parts for text/plain or text/html
	var textPlain, textHTML string
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err == nil {
				textPlain = string(decoded)
			}
		} else if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err == nil {
				textHTML = string(decoded)
			}
		}

		// Recursively check nested parts
		if len(part.Parts) > 0 {
			nestedBody, _ := c.extractBody(part)
			if nestedBody != "" && textPlain == "" {
				textPlain = nestedBody
			}
		}
	}

	// Prefer text/plain over text/html
	if textPlain != "" {
		return textPlain, nil
	}
	if textHTML != "" {
		return textHTML, nil
	}

	return "", fmt.Errorf("no body found")
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

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
