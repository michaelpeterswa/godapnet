package godapnet

import (
	"errors"
	"fmt"
)

var (
	// ErrNilConfig is returned when Send is called without a MessageConfig.
	ErrNilConfig = errors.New("message config is nil")

	// ErrNoCallsigns is returned when a MessageConfig has no destination
	// callsigns.
	ErrNoCallsigns = errors.New("no destination callsigns")

	// ErrInvalidMaxLength is returned when the maximum message length is not
	// positive.
	ErrInvalidMaxLength = errors.New("max message length must be positive")

	// ErrPrefixTooLong is returned when the prefix, plus its ": " separator,
	// leaves no room for text within the maximum message length.
	ErrPrefixTooLong = errors.New("prefix leaves no room for text")

	// ErrEmptyText is returned when the text to send is empty or only
	// whitespace.
	ErrEmptyText = errors.New("text is empty")
)

// StatusError is returned when DAPNET answers a call with anything other than
// 201 Created. Use errors.As to inspect it, for example to detect rejected
// credentials with StatusCode == http.StatusUnauthorized.
type StatusError struct {
	StatusCode int
	Status     string
	// Body is the start of DAPNET's response body, trimmed of whitespace.
	Body string
}

func (e *StatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unexpected response: %s", e.Status)
	}
	return fmt.Sprintf("unexpected response: %s: %s", e.Status, e.Body)
}
