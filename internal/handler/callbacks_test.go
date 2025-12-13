package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanCallbackData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "test_data",
			expected: "test_data",
		},
		{
			name:     "string with whitespace",
			input:    "  test_data  ",
			expected: "test_data",
		},
		{
			name:     "string with newline",
			input:    "test\ndata",
			expected: "testdata",
		},
		{
			name:     "string with tab",
			input:    "test\tdata",
			expected: "testdata",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "string with unprintable characters",
			input:    "test\x00data\x01",
			expected: "testdata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanCallbackData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
