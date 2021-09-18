package godapnet

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func CreateMessage(text string, callsignNames []string, transmitterGroups []string, emergency bool) []Message {
	var messages []Message
	texts := sliceStringByN(text, MaxMessageLength)

	for _, msg := range texts {
		currentMessage := Message{
			Text:              msg,
			CallsignNames:     callsignNames,
			TransmitterGroups: transmitterGroups,
			Emergency:         emergency,
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
			log.Printf("generatePayload() Failed: %s", err.Error())
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
			log.Printf("New Request Failed: %s", err.Error())
		}

		req.Header.Add("Authorization", createAuthToken(username, password))
		req.Header.Set("Content-Type", "application/json")

		_, err = client.Do(req)
		if err != nil {
			log.Printf("Send Request Failed: %s", err.Error())
		}
	}

}
