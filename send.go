package godapnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Sender struct {
	client *http.Client

	URL      string
	Callsign string
	Username string
	Password string
}

func NewSender(client *http.Client, url string, callsign string, username string, password string) *Sender {
	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 30,
		}
	}

	return &Sender{
		client:   client,
		URL:      url,
		Callsign: callsign,
		Username: username,
		Password: password,
	}
}

type MessageConfig struct {
	MaxMessageLength  int
	Prefix            string
	Callsigns         []string
	TransmitterGroups []string
	Emergency         bool
}

func NewMessageConfig(prefix string, maxMessageLength int, callsigns []string, transmitterGroups []string, emergency bool) *MessageConfig {
	return &MessageConfig{
		Prefix:            prefix,
		MaxMessageLength:  maxMessageLength,
		Callsigns:         callsigns,
		TransmitterGroups: transmitterGroups,
		Emergency:         emergency,
	}
}

type Message struct {
	Text                  string   `json:"text"`
	CallsignNames         []string `json:"callSignNames"`
	TransmitterGroupNames []string `json:"transmitterGroupNames"`
	Emergency             bool     `json:"emergency"`
}

func (mc *MessageConfig) createMessage(text string) Message {
	return Message{
		Text:                  text,
		CallsignNames:         mc.Callsigns,
		TransmitterGroupNames: mc.TransmitterGroups,
		Emergency:             mc.Emergency,
	}
}

// Send sends a message to the DAPNET network
func (s *Sender) Send(text string, messageConfig *MessageConfig) error {
	texts := splitText(text, messageConfig.Prefix, messageConfig.MaxMessageLength)

	reversedTexts := make([]string, len(texts))
	for i, n := range texts {
		reversedTexts[len(texts)-1-i] = n
	}

	for _, message := range reversedTexts {
		messageToSend := messageConfig.createMessage(message)
		err := s.sendMessage(messageToSend)
		if err != nil {
			return err
		}
	}

	return nil
}

// splitText splits a text into multiple texts, each with a maximum length
func splitText(inputText string, prefix string, maxLength int) []string {
	var length int
	// Calculate length of message, including "xxxxx: " for prefix at start of message
	if prefix == "" {
		length = maxLength
	} else {
		length = maxLength - len(prefix) - 2
	}

	texts := sliceStringByN(inputText, length)
	prefixedTexts := prefixTexts(texts, prefix)

	return prefixedTexts
}

// prefixTests adds a prefix to each text
func prefixTexts(texts []string, prefix string) []string {
	if prefix == "" {
		return texts
	}

	var prefixedTexts []string
	for _, text := range texts {
		prefixedTexts = append(prefixedTexts, fmt.Sprintf("%s: %s", prefix, text))
	}

	return prefixedTexts
}

func (s *Sender) sendMessage(message Message) error {
	// create json writer
	out, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewBuffer(out))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.SetBasicAuth(s.Username, s.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		_, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
	} else {
		return fmt.Errorf("error sending message: %s", resp.Status)
	}
	return nil
}
