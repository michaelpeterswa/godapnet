package godapnet_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"nw.codes/godapnet"
)

func mockDAPNetCallsEndpoint(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var message godapnet.Message

	err := json.Unmarshal(body, &message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(message.Text) > 80 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func TestSend(t *testing.T) {
	tests := []struct {
		// test config
		name string

		// server config
		serverCallsign string
		serverUsername string
		serverPassword string

		// message config
		mcCallsigns []string
		mcTxGroups  []string
		mcEmergency bool

		// message text
		message string
	}{
		{
			name:           "send test 01",
			serverCallsign: "x1xxx",
			serverUsername: "x1xxx",
			serverPassword: "password",
			mcCallsigns:    []string{"x1xxx"},
			mcTxGroups:     []string{"us-wa"},
			mcEmergency:    false,
			message:        "test message test message test message",
		},
		{
			name:           "send test 02",
			serverCallsign: "x1xxx",
			serverUsername: "x1xxx",
			serverPassword: "password",
			mcCallsigns:    []string{"x1xxx"},
			mcTxGroups:     []string{"us-wa"},
			mcEmergency:    false,
			message:        "this is a test message that is longer than 80 characters and it's important that it gets split up into multiple messages",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mockDAPNetCallsEndpoint(w, r)
			}))
			defer server.Close()

			sender := godapnet.NewSender(http.DefaultClient, server.URL, tc.serverCallsign, tc.serverUsername, tc.serverPassword)

			messageConfig := godapnet.NewMessageConfig(godapnet.Alphapoc602RMaxMessageLength, tc.mcCallsigns, tc.mcTxGroups, tc.mcEmergency)

			err := sender.Send(tc.message, messageConfig)
			if err != nil {
				assert.Fail(t, err.Error())
			}
		},
		)
	}
}
