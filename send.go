package godapnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"
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

func GeneratePayload(messages []Message) []string {
	var payloads []string
	for _, message := range messages {
		payload, err := json.Marshal(message)
		if err != nil {
			logger.Error("Payload failed to marshal", zap.String("error", err.Error()))
		}
		payloads = append(payloads, string(payload))
	}

	return payloads
}

func SendMessage(payloads []string, username string, password string) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	for _, message := range payloads {
		req, err := http.NewRequest("POST", BaseURL+CallsEndpoint, bytes.NewBuffer([]byte(message)))
		if err != nil {
			logger.Error("http.NewRequest failed", zap.String("error", err.Error()))
		}

		req.Header.Add("Authorization", createAuthToken(username, password))
		req.Header.Set("Content-Type", "application/json")

		logger.Debug("Sending a Request",
			zap.String("method", req.Method),
			zap.String("host", req.Host),
		)
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Sending Request Failed", zap.String("error", err.Error()))
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			_, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Reading Response Body Failed", zap.String("error", err.Error()))
			}
			logger.Info("Response: Successful", zap.String("status", resp.Status))
		} else {
			logger.Info("Response: Failed", zap.String("status", resp.Status))
		}

	}

}
