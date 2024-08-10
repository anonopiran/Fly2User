package config

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDotEnv(t *testing.T) {
	tests := []struct {
		name       string
		envContent string
		envKey     []string
		envValue   []string
	}{
		{
			name:       "No .env file",
			envContent: "",
		},
		{
			name:       ".env file with valid content",
			envContent: "KEY=VALUE\n",
			envKey:     []string{"KEY"},
			envValue:   []string{"VALUE"},
		},
		{
			name:       ".env file with multiple variables",
			envContent: "KEY1=VALUE1\nKEY2=VALUE2\n",
			envKey:     []string{"KEY1", "KEY2"},
			envValue:   []string{"VALUE1", "VALUE2"},
		},
	}
	// ...
	f, err := os.MkdirTemp("", "envtest")
	if err != nil {
		panic(err)
	}
	oldPath := os.Getenv("path")
	os.Setenv("path", f)
	defer func() { os.Setenv("path", oldPath) }()
	// ...
	// ...
	for _, tt := range tests {
		if tt.envContent != "" {
			os.Clearenv()
			file, err := os.Create(".env")
			if err != nil {
				panic(err)
			}
			_, err = file.WriteString(tt.envContent)
			file.Close()
			if err != nil {
				panic(err)
			}
		}
		// ...
		os.Clearenv()
		LoadDotEnv()
		if tt.envKey != nil {
			for c, k := range tt.envKey {
				value, exists := os.LookupEnv(k)
				assert.True(t, exists)
				assert.Equal(t, tt.envValue[c], value)
			}

		}
		// ...
		if tt.envContent != "" {
			os.Remove(".env")
		}
	}
}

// ...
func TestDoUnmarshal(t *testing.T) {
	up_urls := []url.URL{}
	for _, u := range []string{"grpc://xray@test.com:123", "grpc://v2fly@test2.com:124"} {
		ur, _ := url.Parse(u)
		up_urls = append(up_urls, *ur)
	}
	tests := []struct {
		name      string
		envVars   map[string]string
		expected  ConfigType
		expectErr bool
	}{
		{
			name:    "Default configuration",
			envVars: map[string]string{},
			expected: ConfigType{
				LogLevel: "WARNING",
				Server: ServerConfigType{
					Listen: ":3000",
				},
				Supervisor: SupervisorConfigType{
					Interval: 60,
				},
			},
			expectErr: false,
		},
		{
			name: "Custom conf",
			envVars: map[string]string{
				"LOG_LEVEL":              "DEBUG",
				"SUPERVISOR__INTERVAL":   "80",
				"SERVER__AUTH":           "user:pass",
				"SERVER__LISTEN":         ":444",
				"SERVER__USER_DB":        ":444",
				"UPSTREAM__ADDRESS":      "grpc://xray@test.com:123,grpc://v2fly@test2.com:124",
				"UPSTREAM__INBOUND_LIST": "i1:vless,i2:vmess",
			},
			expected: ConfigType{
				LogLevel: "DEBUG",
				Server: ServerConfigType{
					Listen: ":444",
					Auth:   AuthType{Username: "user", Password: "pass"},
				},
				Supervisor: SupervisorConfigType{
					Interval: 80,
					UserDB:   "path/to/db.sqlite",
				},
				Upstream: UpstreamConfigType{
					Address: []UpstreamUrlType{{
						URL:        up_urls[0],
						ServerType: XRAY_SRV,
					}, {
						URL:        up_urls[1],
						ServerType: V2FLY_SRV,
					}},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			defer func() {
				// Unset environment variables after the test
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			var cfg ConfigType
			err := doUnmarshal(&cfg)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.LogLevel, cfg.LogLevel)
				assert.Equal(t, tt.expected.Server.Listen, cfg.Server.Listen)
				assert.Equal(t, tt.expected.Supervisor.Interval, cfg.Supervisor.Interval)
			}
		})
	}
}

// ...
func TestDoValidate(t *testing.T) {
	_upUrls := []url.URL{}
	for _, u := range []string{"grpc://xray@test.com:123", "grpc://v2fly@test2.com:124"} {
		ur, _ := url.Parse(u)
		_upUrls = append(_upUrls, *ur)
	}
	_dfltServer := ServerConfigType{
		Listen: ":8080",
		Auth: AuthType{
			Username: "user",
			Password: "pass",
		},
	}
	_dfltSuper := SupervisorConfigType{
		Interval: 30,
		UserDB:   "path/to/user/db",
	}
	_dfltUpstream := UpstreamConfigType{
		Address: []UpstreamUrlType{{
			URL:        _upUrls[0],
			ServerType: XRAY_SRV,
		}, {
			URL:        _upUrls[1],
			ServerType: V2FLY_SRV,
		}},
		InboundList: []InboundConfigType{{Tag: "t1", Proto: VMESS_PROTO}, {Tag: "t2", Proto: VLESS_PROTO}, {Tag: "t3", Proto: TROJAN_PROTO}},
	}
	tests := []struct {
		name      string
		config    ConfigType
		expectErr bool
	}{
		{
			name: "Valid configuration",
			config: ConfigType{
				LogLevel:   "INFO",
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
				Upstream:   _dfltUpstream,
			},
			expectErr: false,
		},
		{
			name: "Missing server listen",
			config: ConfigType{
				LogLevel: "INFO",
				Server: ServerConfigType{
					Auth: _dfltServer.Auth,
				},
				Supervisor: _dfltSuper,
				Upstream:   _dfltUpstream,
			},
			expectErr: true,
		},
		{
			name: "Missing server auth",
			config: ConfigType{
				LogLevel: "INFO",
				Server: ServerConfigType{
					Listen: _dfltServer.Listen,
				},
				Supervisor: _dfltSuper,
				Upstream:   _dfltUpstream,
			},
			expectErr: true,
		},
		{
			name: "Missing supervisor interval",
			config: ConfigType{
				LogLevel: "INFO",
				Server:   _dfltServer,
				Supervisor: SupervisorConfigType{
					UserDB: _dfltSuper.UserDB,
				},
				Upstream: _dfltUpstream,
			},
			expectErr: true,
		}, {
			name: "Missing supervisor db path",
			config: ConfigType{
				LogLevel: "INFO",
				Server:   _dfltServer,
				Supervisor: SupervisorConfigType{
					Interval: _dfltSuper.Interval,
				},
				Upstream: _dfltUpstream,
			},
			expectErr: true,
		}, {
			name: "Missing server",
			config: ConfigType{
				LogLevel:   "INFO",
				Supervisor: _dfltSuper,
				Upstream:   _dfltUpstream,
			},

			expectErr: true,
		}, {
			name: "Missing supervisor",
			config: ConfigType{
				LogLevel: "INFO",
				Server:   _dfltServer,
				Upstream: _dfltUpstream,
			},
			expectErr: true,
		}, {
			name: "Missing log level",
			config: ConfigType{
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
				Upstream:   _dfltUpstream,
			},
			expectErr: true,
		}, {
			name: "Empty upstream",
			config: ConfigType{
				LogLevel:   "INFO",
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
				Upstream:   UpstreamConfigType{},
			},
			expectErr: true,
		}, {
			name: "Empty upstream arddr",
			config: ConfigType{
				LogLevel:   "INFO",
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
				Upstream: UpstreamConfigType{
					Address:     []UpstreamUrlType{},
					InboundList: _dfltUpstream.InboundList,
				},
			},
			expectErr: true,
		},
		{
			name: "Empty upstream inbound",
			config: ConfigType{
				LogLevel:   "INFO",
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
				Upstream: UpstreamConfigType{
					Address:     _dfltUpstream.Address,
					InboundList: []InboundConfigType{},
				},
			},
			expectErr: true,
		}, {
			name: "Missing upstream arddr",
			config: ConfigType{
				LogLevel:   "INFO",
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
				Upstream: UpstreamConfigType{
					InboundList: _dfltUpstream.InboundList,
				},
			},
			expectErr: true,
		},
		{
			name: "Missing upstream inbound",
			config: ConfigType{
				LogLevel:   "INFO",
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
				Upstream: UpstreamConfigType{
					Address: _dfltUpstream.Address,
				},
			},
			expectErr: true,
		}, {
			name: "Missing upstream",
			config: ConfigType{
				LogLevel:   "INFO",
				Server:     _dfltServer,
				Supervisor: _dfltSuper,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := doValidate(&tt.config)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
