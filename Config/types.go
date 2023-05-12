package config

import (
	u2u "Fly2User/User2Upstream"
	"errors"
	"net/url"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type PathType string
type LogLevelType string
type V2rayUrlType string
type UUIDType string
type SettingsType struct {
	V2flyApiAddress   []V2rayUrlType    `env:"V2FLY_API_ADDRESS,required"`
	SuperviseInterval int               `env:"SUPERVISE_INTERVAL" envDefault:"5"`
	SuperviseUuid     UUIDType          `env:"SUPERVISE_UUID,required"`
	UserDir           PathType          `env:"USER_DIR" envDefault:"./storage/users"`
	InboundList       []u2u.InboundType `env:"INBOUND_LIST,required"`
	LogLevel          LogLevelType      `env:"LOG_LEVEL" envDefault:"warning"`
	Listen            string            `env:"LISTEN" envDefault:":3000"`
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
	LogWithRaw := log.WithField("value", string(s))
	err := os.MkdirAll(s, os.ModePerm)
	if err != nil {
		LogWithRaw.Error(err)
		return err
	}
	*f = PathType(s)
	return nil
}
func (f *V2rayUrlType) UnmarshalText(text []byte) error {
	s := string(text)
	LogWithRaw := log.WithField("value", s)
	u, err := url.Parse(s)
	if err != nil {
		LogWithRaw.Error(err)
		return err
	}
	if u.Hostname() == "" {
		e := errors.New("hostname not provided")
		LogWithRaw.Error(e)
		return e
	}
	if u.Port() == "" {
		e := errors.New("port not provided")
		LogWithRaw.Error(e)
		return e
	}
	*f = V2rayUrlType(s)
	return nil
}
func (f *UUIDType) UnmarshalText(text []byte) error {
	s := string(text)
	LogWithRaw := log.WithField("value", s)
	_, err := uuid.Parse(s)
	if err != nil {
		LogWithRaw.Error(err)
		return err
	}
	*f = UUIDType(s)
	return nil
}
func (f *LogLevelType) UnmarshalText(text []byte) error {
	s := string(text)
	LogWithRaw := log.WithField("value", s)
	ll, err := log.ParseLevel(s)
	if err != nil {
		LogWithRaw.Error(err)
		return err
	}
	log.SetLevel(ll)
	*f = LogLevelType(s)
	return nil
}
