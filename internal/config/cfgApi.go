package config

import (
	"errors"
	"strings"
)

type AuthType struct {
	Username string `koanf:"username" validate:"required"`
	Password string `koanf:"password" validate:"required"`
}
type ServerConfigType struct {
	Auth   AuthType `koanf:"auth" validate:"required"`
	Listen string   `koanf:"listen" validate:"required"`
}

// ...
func (f *AuthType) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return nil
	}
	splt := strings.SplitN(s, ":", 2)
	if len(splt) < 2 {
		return errors.New("user:pass not understoord")
	}
	*f = AuthType{
		Username: splt[0],
		Password: splt[1],
	}
	return nil
}
