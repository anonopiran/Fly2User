package config

import (
	u2u "Fly2User/User2Upstream"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type PathType string
type LogLevelType string
type V2rayUrlType string
type UUIDType string
type AuthType struct {
	Username string
	Password string
}
type SettingsType struct {
	V2flyApiAddress   []V2rayUrlType    `env:"V2FLY_API_ADDRESS,required"`
	SuperviseInterval int               `env:"SUPERVISE_INTERVAL" envDefault:"5"`
	SuperviseUuid     UUIDType          `env:"SUPERVISE_UUID,required"`
	UserDir           PathType          `env:"USER_DIR" envDefault:"./storage/users"`
	InboundList       []u2u.InboundType `env:"INBOUND_LIST,required"`
	LogLevel          LogLevelType      `env:"LOG_LEVEL" envDefault:"warning"`
	Listen            string            `env:"LISTEN" envDefault:":3000"`
	Auth              []AuthType        `env:"AUTH" envDefault:""`
}

func (f *PathType) AsString() string {
	return string(*f)
}
func (f *V2rayUrlType) AsUrl() url.URL {
	u, _ := url.Parse(string(*f))
	return *u
}
func (f *UUIDType) AsString() string {
	return string(*f)

}

// .....
func (f *PathType) UnmarshalText(text []byte) error {
	s := string(text)
	err := os.MkdirAll(s, os.ModePerm)
	if err != nil {
		return err
	}
	*f = PathType(s)
	return nil
}
func (f *V2rayUrlType) UnmarshalText(text []byte) error {
	s := string(text)
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u.Hostname() == "" {
		e := errors.New("hostname not provided")
		return e
	}
	if u.Port() == "" {
		e := errors.New("port not provided")
		return e
	}
	*f = V2rayUrlType(s)
	return nil
}
func (f *UUIDType) UnmarshalText(text []byte) error {
	s := string(text)
	_, err := uuid.Parse(s)
	if err != nil {
		return err
	}
	*f = UUIDType(s)
	return nil
}
func (f *LogLevelType) UnmarshalText(text []byte) error {
	s := string(text)
	ll, err := log.ParseLevel(s)
	if err != nil {
		return err
	}
	log.SetLevel(ll)
	*f = LogLevelType(s)
	return nil
}
func (f *AuthType) UnmarshalText(text []byte) error {
	s := string(text)
	splt := strings.SplitN(s, ":", 2)
	if len(splt) < 2 {
		return fmt.Errorf("not user/pass detected in %s", s)
	}
	*f = AuthType{
		Username: splt[0],
		Password: splt[1],
	}
	return nil
}
