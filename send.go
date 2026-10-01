// Package godapnet sends messages to amateur radio POCSAG pagers over the
// DAPNET network (https://hampager.de).
package godapnet

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

// maxErrorBodySize caps how much of an error response is kept in a StatusError.
const maxErrorBodySize = 512

// Sender sends calls to DAPNET with one account's credentials.
type Sender struct {
	client   *http.Client
	url      string
	username string
	password string
}

// SenderOption configures a Sender.
type SenderOption func(*Sender)

// NewSender returns a Sender that authenticates as username. By default it
// posts to DAPNetURL with an http.Client that times out after 30 seconds.
func NewSender(username string, password string, opts ...SenderOption) *Sender {
	s := &Sender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		url:      DAPNetURL,
		username: username,
		password: password,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithHTTPClient sends requests with client instead of the default.
func WithHTTPClient(client *http.Client) SenderOption {
	return func(s *Sender) {
		s.client = client
	}
}

// WithURL posts calls to url instead of DAPNetURL.
func WithURL(url string) SenderOption {
	return func(s *Sender) {
		s.url = url
	}
}

// Send sends text as one or more pages, as described by messageConfig. Pages
// are sent one at a time, so on error the returned message says how many were
// already delivered.
func (s *Sender) Send(ctx context.Context, text string, messageConfig *MessageConfig) error {
	messages, err := messageConfig.messages(text)
	if err != nil {
		return err
	}

	for i, message := range messages {
		if err := s.sendMessage(ctx, message); err != nil {
			return fmt.Errorf("sent %d of %d pages: %w", i, len(messages), err)
		}
	}

	return nil
}

func (s *Sender) sendMessage(ctx context.Context, message Message) error {
	out, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(out))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return &StatusError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       strings.TrimSpace(string(body)),
		}
	}

	// Drain the body so the connection can be reused.
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	return nil
}
