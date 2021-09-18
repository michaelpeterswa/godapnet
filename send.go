package godapnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func CreateMessage(prefix string, text string, callsignNames []string, transmitterGroupNames []string, emergency bool) []Message {
	var messages []Message
	// Calculate length of message, including "xxxxx: " for prefix at start of message
	length := MaxMessageLength - len(prefix) - 2
	texts := sliceStringByN(text, length)

	for _, msg := range texts {
		currentMessage := Message{
			Text:                  fmt.Sprintf("%s: %s", prefix, msg),
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
			log.Printf("generatePayload() Failed: %s\n", err.Error())
		}
		payloads = append(payloads, string(payload))
	}

	return payloads
}

func SendMessage(payloads []string, username string, password string) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	for _, message := range payloads {
		req, err := http.NewRequest("POST", BaseURL+CallsEndpoint, bytes.NewBuffer([]byte(message)))
		if err != nil {
			log.Printf("New Request Failed: %s\n", err.Error())
		}

		req.Header.Add("Authorization", createAuthToken(username, password))
		req.Header.Set("Content-Type", "application/json")

		log.Printf("Request: %s - %s :: %s - %s\n", req.Method, req.Host, req.Header, req.Body)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Send Request Failed: %s\n", err.Error())
		}
		log.Printf("Response (%s): %s\n", resp.Status, resp.Body)

		defer resp.Body.Close()
	}

}
