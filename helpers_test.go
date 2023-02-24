package godapnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceStringByN(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		n      int
		output []string
	}{
		{
			name:   "testing sliceStringByN even splits",
			input:  "asdfasdfasdfasdf",
			n:      4,
			output: []string{"asdf", "asdf", "asdf", "asdf"},
		},
		{
			name:   "testing sliceStringByN uneven splits",
			input:  "asdfasdfasdfasdfa",
			n:      4,
			output: []string{"asdf", "asdf", "asdf", "asdf", "a"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.output, sliceStringByN(tc.input, tc.n))
		})
	}
}
