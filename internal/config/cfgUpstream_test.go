package config

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamUrlType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected error
		output   *UpstreamUrlType
	}{
		{
			// Happy path with valid GRPC URL and server type

			name:     "valid_grpc_url",
			input:    []byte("grpc://v2fly@test.server:8080"),
			expected: nil,
			output: &UpstreamUrlType{
				URL: url.URL{
					Scheme: "grpc",
					Host:   "test.server:8080",
				},
				ServerType: V2FLY_SRV,
			},
		}, {
			name:     "valid_grpc_url",
			input:    []byte("grpc://xray@test.server:8080"),
			expected: nil,
			output: &UpstreamUrlType{
				URL: url.URL{
					Scheme: "grpc",
					Host:   "test.server:8080",
				},
				ServerType: XRAY_SRV,
			},
		},
		// Invalid scheme
		{
			name:     "invalid_scheme",
			input:    []byte("http://server.example.com:8080/V2FLY_SRV"),
			expected: errors.New("scheme should be grpc"),
			output:   &UpstreamUrlType{},
		},
		// Missing hostname
		{
			name:     "missing_hostname",
			input:    []byte("grpc:///V2FLY_SRV"),
			expected: errors.New("hostname not provided"),
			output:   &UpstreamUrlType{},
		},
		// Missing port
		{
			name:     "missing_port",
			input:    []byte("grpc://server.example.com/V2FLY_SRV"),
			expected: errors.New("port not provided"),
			output:   &UpstreamUrlType{},
		},
		// Unsupported server type
		{
			name:     "unsupported_servertype",
			input:    []byte("grpc://server.example.com:8080/UNSUPPORTED_SRV"),
			expected: errors.New("servertype not understood"),
			output:   &UpstreamUrlType{},
		}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var f UpstreamUrlType
			err := f.UnmarshalText(tc.input)

			// Assertions using testify
			assert.Equal(t, tc.expected, err)
			assert.Equal(t, tc.output, &f)
		})
	}
}

func TestInboundConfigType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected error
		output   *InboundConfigType
	}{
		// Happy path with valid format
		{
			name:     "valid_tag_proto",
			input:    []byte("mytag:VMESS"),
			expected: nil,
			output: &InboundConfigType{
				Tag:   "mytag",
				Proto: VMESS_PROTO,
			},
		},
		{
			name:     "valid_tag_proto",
			input:    []byte("mytag:VLESS"),
			expected: nil,
			output: &InboundConfigType{
				Tag:   "mytag",
				Proto: VLESS_PROTO,
			},
		},
		{
			name:     "valid_tag_proto",
			input:    []byte("mytag:TROJAN"),
			expected: nil,
			output: &InboundConfigType{
				Tag:   "mytag",
				Proto: TROJAN_PROTO,
			},
		},
		// Empty input
		{
			name:     "empty_input",
			input:    []byte(""),
			expected: errors.New("error parsing inbound"),
			output:   &InboundConfigType{},
		},
		// Missing colon separator
		{
			name:     "missing_colon",
			input:    []byte("mytagVLESS"),
			expected: errors.New("error parsing inbound"),
			output:   &InboundConfigType{},
		},
		// Invalid protocol
		{
			name:     "invalid_protocol",
			input:    []byte("mytag:UNKNOWN"),
			expected: errors.New("protocol not understood"),
			output:   &InboundConfigType{},
		},
		// Extra colon separators
		{
			name:     "extra_colon",
			input:    []byte("mytag:vv:VMESS"), // Ignored extra colon
			expected: nil,
			output: &InboundConfigType{
				Tag:   "mytag:vv",
				Proto: VMESS_PROTO,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var f InboundConfigType
			err := f.UnmarshalText(tc.input)

			// Assertions using testify
			assert.Equal(t, tc.expected, err)
			assert.Equal(t, tc.output, &f)
		})
	}
}
