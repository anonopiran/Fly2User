package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert" // Using testify for better assertions
)

type unmarshalAuthTypeTestCase struct {
	name     string
	input    []byte
	expected error
	output   *AuthType
}

func TestAuthType_UnmarshalText(t *testing.T) {
	tests := []unmarshalAuthTypeTestCase{
		// Happy path with valid input
		{
			name:     "valid_user_pass",
			input:    []byte("username:password"),
			expected: nil,
			output: &AuthType{
				Username: "username",
				Password: "password",
			},
		},
		// Empty input
		{
			name:     "empty_input",
			input:    []byte(""),
			expected: nil,
			output:   &AuthType{},
		},
		// Missing colon separator
		{
			name:     "missing_colon",
			input:    []byte("usernamepassword"),
			expected: errors.New("user:pass not understoord"),
			output:   &AuthType{},
		},
		// Extra colon separators
		{
			name:     "extra_colon",
			input:    []byte("username:pass:word"),
			expected: nil, // Function ignores extra separators
			output: &AuthType{
				Username: "username",
				Password: "pass:word",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var f AuthType
			err := f.UnmarshalText(tc.input)

			// Assertions using testify
			assert.Equal(t, tc.expected, err)
			assert.Equal(t, tc.output, &f)
		})
	}
}
