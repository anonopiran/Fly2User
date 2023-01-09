package config

import (
	u2g "Fly2User/User2Grpc"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

// ==================================
// types
// ==================================
type PathType string
type LogLevelType string
type settingsType struct {
	LogLevel          LogLevelType      `env:"LOG_LEVEL" env-default:"warning"`
	V2flyApiAddress   []string          `env:"V2FLY_API_ADDRESS" env-required:"true"`
	SuperviseInterval int               `env:"SUPERVISE_INTERVAL" env-default:"5"`
	SuperviseUuid     string            `env:"SUPERVISE_UUID" env-required:"true"`
	UserDir           PathType          `env:"USER_DIR" env-default:"./storage/users"`
	InboundList       []u2g.InboundType `env:"INBOUND_LIST" env-required:"true"`
}

// ==================================
// functions
// ==================================
func Config() settingsType {
	var cfg settingsType
	var err error = nil
	if _, err_file := os.Stat(".env"); err_file == nil {
		err = cleanenv.ReadConfig(".env", &cfg)
		log.Info("found .env file")
	} else {
		err = cleanenv.ReadEnv(&cfg)
		log.Info("no .env file found")
	}
	if err != nil {
		log.WithError(err).Fatalln("can not initiate configuration")
	}
	log.WithField("data", fmt.Sprintf("%+v", cfg)).Debug("Parsed Configuration")
	return cfg
}
func Describe() {
	var cfg settingsType
	help, err := cleanenv.GetDescription(&cfg, nil)
	if err != nil {
		log.WithError(err).Panic("can not generate description")
	}
	log.Println(help)
}

func (f *PathType) SetValue(s string) error {
	err := os.MkdirAll(s, os.ModePerm)
	if err != nil {
		log.WithField("path", f).WithError(err).Error("can not create dir")
		return err
	}
	*f = PathType(s)
	return nil
}
func (f *LogLevelType) SetValue(s string) error {
	ll, err := log.ParseLevel(s)
	if err != nil {
		log.WithField("value", f).WithError(err).Error("can not set log level")
		return err
	}
	log.SetLevel(ll)
	*f = LogLevelType(s)
	return nil
}
