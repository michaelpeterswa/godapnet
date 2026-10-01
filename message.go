package godapnet

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// MessageConfig describes who receives a message and how its text is split
// into pages. Build one with NewMessageConfig.
type MessageConfig struct {
	callsigns         []string
	transmitterGroups []string
	prefix            string
	maxMessageLength  int
	emergency         bool
	reverseOrder      bool
}

// MessageOption configures a MessageConfig.
type MessageOption func(*MessageConfig)

// NewMessageConfig returns a config that sends to the given callsigns, with
// pages of at most Alphapoc602RMaxMessageLength characters unless overridden.
func NewMessageConfig(callsigns []string, opts ...MessageOption) *MessageConfig {
	mc := &MessageConfig{
		callsigns:        callsigns,
		maxMessageLength: Alphapoc602RMaxMessageLength,
	}

	for _, opt := range opts {
		opt(mc)
	}

	return mc
}

// WithTransmitterGroups sets the transmitter groups that broadcast the message.
func WithTransmitterGroups(groups ...string) MessageOption {
	return func(mc *MessageConfig) {
		mc.transmitterGroups = groups
	}
}

// WithPrefix starts every page with "<prefix>: ", typically the sender's
// callsign. The prefix counts towards the maximum message length.
func WithPrefix(prefix string) MessageOption {
	return func(mc *MessageConfig) {
		mc.prefix = prefix
	}
}

// WithMaxMessageLength sets the maximum number of characters in each page.
func WithMaxMessageLength(n int) MessageOption {
	return func(mc *MessageConfig) {
		mc.maxMessageLength = n
	}
}

// WithEmergency marks the message as an emergency call.
func WithEmergency() MessageOption {
	return func(mc *MessageConfig) {
		mc.emergency = true
	}
}

// WithReverseOrder sends the pages of a long message last page first, for
// pagers that list the most recent message at the top.
func WithReverseOrder() MessageOption {
	return func(mc *MessageConfig) {
		mc.reverseOrder = true
	}
}

// Message is the JSON body DAPNET's calls endpoint expects.
type Message struct {
	Text                  string   `json:"text"`
	CallsignNames         []string `json:"callSignNames"`
	TransmitterGroupNames []string `json:"transmitterGroupNames"`
	Emergency             bool     `json:"emergency"`
}

// messages validates the config and returns the messages to send for text,
// in send order.
func (mc *MessageConfig) messages(text string) ([]Message, error) {
	if mc == nil {
		return nil, ErrNilConfig
	}
	if len(mc.callsigns) == 0 {
		return nil, ErrNoCallsigns
	}
	if mc.maxMessageLength <= 0 {
		return nil, ErrInvalidMaxLength
	}

	pageLength := mc.maxMessageLength
	if mc.prefix != "" {
		pageLength -= utf8.RuneCountInString(mc.prefix) + len(": ")
	}
	if pageLength <= 0 {
		return nil, ErrPrefixTooLong
	}

	pages := splitText(text, pageLength)
	if len(pages) == 0 {
		return nil, ErrEmptyText
	}
	if mc.reverseOrder {
		slices.Reverse(pages)
	}

	messages := make([]Message, len(pages))
	for i, page := range pages {
		if mc.prefix != "" {
			page = mc.prefix + ": " + page
		}
		messages[i] = Message{
			Text:                  page,
			CallsignNames:         mc.callsigns,
			TransmitterGroupNames: mc.transmitterGroups,
			Emergency:             mc.emergency,
		}
	}

	return messages, nil
}

// splitText breaks text into pages of at most n characters, breaking between
// words where it can. Runs of whitespace collapse to a single space, and a word
// longer than a page is split across pages.
func splitText(text string, n int) []string {
	var (
		pages []string
		page  strings.Builder
		size  int // characters in page, which may differ from its byte length
	)

	flush := func() {
		if size > 0 {
			pages = append(pages, page.String())
			page.Reset()
			size = 0
		}
	}

	for _, word := range strings.Fields(text) {
		runes := []rune(word)

		for len(runes) > n {
			flush()
			pages = append(pages, string(runes[:n]))
			runes = runes[n:]
		}

		if size > 0 && size+1+len(runes) > n {
			flush()
		}
		if size > 0 {
			page.WriteByte(' ')
			size++
		}
		page.WriteString(string(runes))
		size += len(runes)
	}
	flush()

	return pages
}
