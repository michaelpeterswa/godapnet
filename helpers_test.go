package godapnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAuthToken(t *testing.T) {
	tests := []struct {
		name                 string
		username             string
		password             string
		encodedAuthorization string
	}{
		{
			name:                 "testing createAuthToken basic",
			username:             "testing",
			password:             "secret",
			encodedAuthorization: "dGVzdGluZzpzZWNyZXQ=",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, createAuthToken(tc.username, tc.password), tc.encodedAuthorization)
		})
	}
}
