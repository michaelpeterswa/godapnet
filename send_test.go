package godapnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMessage(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		text       string
		callNames  []string
		txGrpNames []string
		emergency  bool
		output     []Message
	}{
		{
			name:       "testing CreateMessage basic",
			prefix:     "myprefix",
			text:       "this is a test message",
			callNames:  []string{"testcall"},
			txGrpNames: []string{"us-wa"},
			emergency:  false,
			output: []Message{
				{
					Text:                  "myprefix: this is a test message",
					CallsignNames:         []string{"testcall"},
					TransmitterGroupNames: []string{"us-wa"},
					Emergency:             false,
				},
			},
		},
		{
			name:       "testing CreateMessage empty prefix",
			prefix:     "",
			text:       "this is a test message",
			callNames:  []string{"testcall"},
			txGrpNames: []string{"us-wa"},
			emergency:  false,
			output: []Message{
				{
					Text:                  "this is a test message",
					CallsignNames:         []string{"testcall"},
					TransmitterGroupNames: []string{"us-wa"},
					Emergency:             false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.output, CreateMessage(tc.prefix, tc.text, tc.callNames, tc.txGrpNames, tc.emergency))
		})
	}
}

func TestGeneratePayloads(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		output   []string
	}{
		{
			name: "testing GeneratePayloads basic",
			messages: []Message{
				{
					Text:                  "myprefix: this is a test message",
					CallsignNames:         []string{"testcall"},
					TransmitterGroupNames: []string{"us-wa"},
					Emergency:             false,
				},
			},
			output: []string{"{\"text\":\"myprefix: this is a test message\",\"callSignNames\":[\"testcall\"],\"transmitterGroupNames\":[\"us-wa\"],\"emergency\":false}"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := GeneratePayload(tc.messages)
			if err != nil {
				t.FailNow()
			}
			assert.Equal(t, tc.output, res)
		})
	}
}
