package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

var Config SettingsType

func Describe() {
	var cfg SettingsType
	help, err := cleanenv.GetDescription(&cfg, nil)
	if err != nil {
		log.WithError(err).Panic("can not generate description")
	}
	log.Println(help)
}

func init() {
	var err error = nil
	if _, err_file := os.Stat(".env"); err_file == nil {
		err = cleanenv.ReadConfig(".env", &Config)
		log.Info("found .env file")
	} else {
		err = cleanenv.ReadEnv(&Config)
		log.Info("no .env file found")
	}
	if err != nil {
		log.WithError(err).Panic("can not initiate configuration")
	}
	log.WithField("data", fmt.Sprintf("%+v", Config)).Debug("Parsed Configuration")
}
