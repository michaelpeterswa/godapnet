package godapnet_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/michaelpeterswa/godapnet"
)

// fakeDAPNet records every call it receives. It answers with status, or 201
// Created when status is zero.
type fakeDAPNet struct {
	t      *testing.T
	status int
	body   string

	mu    sync.Mutex
	calls []godapnet.Message
}

func (f *fakeDAPNet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assert.Equal(f.t, http.MethodPost, r.Method)
	assert.Equal(f.t, "application/json", r.Header.Get("Content-Type"))

	username, password, ok := r.BasicAuth()
	assert.True(f.t, ok, "request has no basic auth")
	assert.Equal(f.t, "x1xxx", username)
	assert.Equal(f.t, "password", password)

	var message godapnet.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f.mu.Lock()
	f.calls = append(f.calls, message)
	f.mu.Unlock()

	if f.status != 0 {
		w.WriteHeader(f.status)
		_, _ = w.Write([]byte(f.body))
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (f *fakeDAPNet) texts() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	var texts []string
	for _, call := range f.calls {
		texts = append(texts, call.Text)
	}
	return texts
}

func newSender(t *testing.T, fake *fakeDAPNet) *godapnet.Sender {
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	return godapnet.NewSender("x1xxx", "password",
		godapnet.WithHTTPClient(server.Client()),
		godapnet.WithURL(server.URL),
	)
}

func TestSend(t *testing.T) {
	tests := []struct {
		name  string
		opts  []godapnet.MessageOption
		text  string
		texts []string
	}{
		{
			name:  "single page",
			opts:  []godapnet.MessageOption{godapnet.WithPrefix("x1xxx")},
			text:  "test message test message test message",
			texts: []string{"x1xxx: test message test message test message"},
		},
		{
			name: "splits long text, last page first",
			opts: []godapnet.MessageOption{godapnet.WithPrefix("x1xxx")},
			text: "this is a test message that is longer than 80 characters and it's important that it gets split up into multiple messages",
			texts: []string{
				"x1xxx: important that it gets split up into multiple messages",
				"x1xxx: this is a test message that is longer than 80 characters and it's",
			},
		},
		{
			name: "splits long text in order",
			opts: []godapnet.MessageOption{godapnet.WithMaxMessageLength(10), godapnet.WithInOrder()},
			text: "first page second page",
			texts: []string{
				"first page",
				"second",
				"page",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeDAPNet{t: t}
			sender := newSender(t, fake)

			opts := append([]godapnet.MessageOption{godapnet.WithTransmitterGroups("us-wa")}, tc.opts...)
			config := godapnet.NewMessageConfig([]string{"x1xxx"}, opts...)

			require.NoError(t, sender.Send(context.Background(), tc.text, config))
			assert.Equal(t, tc.texts, fake.texts())

			for _, call := range fake.calls {
				assert.Equal(t, []string{"x1xxx"}, call.CallsignNames)
				assert.Equal(t, []string{"us-wa"}, call.TransmitterGroupNames)
				assert.LessOrEqual(t, len([]rune(call.Text)), godapnet.Alphapoc602RMaxMessageLength)
			}
		})
	}
}

func TestSendStatusError(t *testing.T) {
	fake := &fakeDAPNet{t: t, status: http.StatusUnauthorized, body: "  {\"message\":\"bad credentials\"}\n"}
	sender := newSender(t, fake)
	config := godapnet.NewMessageConfig([]string{"x1xxx"}, godapnet.WithMaxMessageLength(5))

	err := sender.Send(context.Background(), "one two", config)
	require.Error(t, err)

	var statusErr *godapnet.StatusError
	require.True(t, errors.As(err, &statusErr), "error %v is not a *StatusError", err)
	assert.Equal(t, http.StatusUnauthorized, statusErr.StatusCode)
	assert.Equal(t, `{"message":"bad credentials"}`, statusErr.Body)
	assert.Contains(t, err.Error(), "sent 0 of 2 pages")

	// The first failure stops the rest of the pages being sent. Pages go last
	// page first, so only "two" reached the server.
	assert.Equal(t, []string{"two"}, fake.texts())
}

func TestSendInvalidConfig(t *testing.T) {
	fake := &fakeDAPNet{t: t}
	sender := newSender(t, fake)

	err := sender.Send(context.Background(), "", godapnet.NewMessageConfig([]string{"x1xxx"}))
	assert.ErrorIs(t, err, godapnet.ErrEmptyText)
	assert.Empty(t, fake.texts(), "nothing should be sent for an invalid message")
}

func TestSendCanceledContext(t *testing.T) {
	fake := &fakeDAPNet{t: t}
	sender := newSender(t, fake)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.Send(ctx, "hello", godapnet.NewMessageConfig([]string{"x1xxx"}))
	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, fake.texts())
}
