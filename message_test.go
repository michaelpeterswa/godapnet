package godapnet

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		n     int
		pages []string
	}{
		{
			name:  "fits on one page",
			text:  "hello there",
			n:     20,
			pages: []string{"hello there"},
		},
		{
			name:  "exactly one page",
			text:  "hello there",
			n:     11,
			pages: []string{"hello there"},
		},
		{
			name:  "breaks between words",
			text:  "the quick brown fox jumps",
			n:     10,
			pages: []string{"the quick", "brown fox", "jumps"},
		},
		{
			name:  "collapses spaces and tabs",
			text:  "  one\t\ttwo   three  ",
			n:     20,
			pages: []string{"one two three"},
		},
		{
			name:  "keeps line breaks",
			text:  "one two\n three  \nfour",
			n:     20,
			pages: []string{"one two\nthree\nfour"},
		},
		{
			name:  "keeps blank lines",
			text:  "one\n\ntwo",
			n:     20,
			pages: []string{"one\n\ntwo"},
		},
		{
			name:  "treats CRLF as a line break",
			text:  "one\r\ntwo",
			n:     20,
			pages: []string{"one\ntwo"},
		},
		{
			name:  "drops leading and trailing line breaks",
			text:  "\n\n one\ntwo \n\n",
			n:     20,
			pages: []string{"one\ntwo"},
		},
		{
			name:  "counts a line break as a character",
			text:  "abcd\nefgh",
			n:     9,
			pages: []string{"abcd\nefgh"},
		},
		{
			name:  "drops a line break where a page ends",
			text:  "abcd\nefgh",
			n:     8,
			pages: []string{"abcd", "efgh"},
		},
		{
			name:  "line break after a split word",
			text:  "abcdefgh\nij",
			n:     6,
			pages: []string{"abcdef", "gh\nij"},
		},
		{
			name:  "splits a word longer than a page",
			text:  "hi abcdefghijklmnop ok",
			n:     6,
			pages: []string{"hi", "abcdef", "ghijkl", "mnop", "ok"},
		},
		{
			name:  "counts characters not bytes",
			text:  "ünïcödé wörds",
			n:     7,
			pages: []string{"ünïcödé", "wörds"},
		},
		{
			name:  "never splits inside a character",
			text:  "ääääääää",
			n:     3,
			pages: []string{"äää", "äää", "ää"},
		},
		{
			name:  "empty",
			text:  "",
			n:     10,
			pages: nil,
		},
		{
			name:  "only whitespace",
			text:  " \t\n\r\n ",
			n:     10,
			pages: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pages := splitText(tc.text, tc.n)
			assert.Equal(t, tc.pages, pages)
			for _, page := range pages {
				assert.True(t, utf8.ValidString(page), "page %q is not valid UTF-8", page)
				assert.LessOrEqual(t, utf8.RuneCountInString(page), tc.n)
			}
		})
	}
}

func TestMessages(t *testing.T) {
	t.Run("prefixes every page within the max length", func(t *testing.T) {
		mc := NewMessageConfig([]string{"x1xxx"}, WithPrefix("x1xxx"), WithMaxMessageLength(16), WithInOrder())

		messages, err := mc.messages("one two three four")
		require.NoError(t, err)

		var texts []string
		for _, m := range messages {
			texts = append(texts, m.Text)
			assert.LessOrEqual(t, utf8.RuneCountInString(m.Text), 16)
		}
		assert.Equal(t, []string{"x1xxx: one two", "x1xxx: three", "x1xxx: four"}, texts)
	})

	t.Run("sends the last page first by default", func(t *testing.T) {
		mc := NewMessageConfig([]string{"x1xxx"}, WithMaxMessageLength(5))

		messages, err := mc.messages("one two three")
		require.NoError(t, err)

		var texts []string
		for _, m := range messages {
			texts = append(texts, m.Text)
		}
		assert.Equal(t, []string{"three", "two", "one"}, texts)
	})

	t.Run("in order", func(t *testing.T) {
		mc := NewMessageConfig([]string{"x1xxx"}, WithMaxMessageLength(5), WithInOrder())

		messages, err := mc.messages("one two three")
		require.NoError(t, err)

		var texts []string
		for _, m := range messages {
			texts = append(texts, m.Text)
		}
		assert.Equal(t, []string{"one", "two", "three"}, texts)
	})

	t.Run("copies recipients and emergency flag", func(t *testing.T) {
		mc := NewMessageConfig(
			[]string{"x1xxx", "x2xxx"},
			WithTransmitterGroups("us-wa", "us-or"),
			WithEmergency(),
		)

		messages, err := mc.messages("hello")
		require.NoError(t, err)
		assert.Equal(t, []Message{{
			Text:                  "hello",
			CallsignNames:         []string{"x1xxx", "x2xxx"},
			TransmitterGroupNames: []string{"us-wa", "us-or"},
			Emergency:             true,
		}}, messages)
	})

	errorTests := []struct {
		name string
		mc   *MessageConfig
		text string
		err  error
	}{
		{
			name: "nil config",
			mc:   nil,
			text: "hello",
			err:  ErrNilConfig,
		},
		{
			name: "zero value config",
			mc:   &MessageConfig{},
			text: "hello",
			err:  ErrNoCallsigns,
		},
		{
			name: "no callsigns",
			mc:   NewMessageConfig(nil),
			text: "hello",
			err:  ErrNoCallsigns,
		},
		{
			name: "zero max length",
			mc:   NewMessageConfig([]string{"x1xxx"}, WithMaxMessageLength(0)),
			text: "hello",
			err:  ErrInvalidMaxLength,
		},
		{
			name: "prefix fills the page",
			mc:   NewMessageConfig([]string{"x1xxx"}, WithPrefix("x1xxx"), WithMaxMessageLength(7)),
			text: "hello",
			err:  ErrPrefixTooLong,
		},
		{
			name: "prefix longer than the page",
			mc:   NewMessageConfig([]string{"x1xxx"}, WithPrefix("W1AWLONGPREFIX"), WithMaxMessageLength(10)),
			text: "hello there",
			err:  ErrPrefixTooLong,
		},
		{
			name: "empty text",
			mc:   NewMessageConfig([]string{"x1xxx"}),
			text: "   ",
			err:  ErrEmptyText,
		},
	}

	for _, tc := range errorTests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.mc.messages(tc.text)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
