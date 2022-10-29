package godapnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func CreateMessage(prefix string, text string, callsignNames []string, transmitterGroupNames []string, emergency bool) []Message {
	var messages []Message
	var length int
	// Calculate length of message, including "xxxxx: " for prefix at start of message
	if prefix == "" {
		length = MaxMessageLength
	} else {
		length = MaxMessageLength - len(prefix) - 2
	}

	texts := sliceStringByN(text, length)

	for _, msg := range texts {
		var text string
		if length == MaxMessageLength {
			text = msg
		} else {
			text = fmt.Sprintf("%s: %s", prefix, msg)
		}
		currentMessage := Message{
			Text:                  text,
			CallsignNames:         callsignNames,
			TransmitterGroupNames: transmitterGroupNames,
			Emergency:             emergency,
		}
		messages = append(messages, currentMessage)
	}

	return messages
}

func GeneratePayload(messages []Message) ([]string, error) {
	var payloads []string
	for _, message := range messages {
		payload, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, string(payload))
	}

	return payloads, nil
}

func SendMessage(payloads []string, username string, password string) error {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	for _, message := range payloads {
		req, err := http.NewRequest("POST", BaseURL+CallsEndpoint, bytes.NewBuffer([]byte(message)))
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Add("Authorization", createAuthToken(username, password))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
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
	}
	return nil
}
