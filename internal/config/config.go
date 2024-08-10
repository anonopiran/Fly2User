package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/sirupsen/logrus"
)

// ...
type ConfigType struct {
	Server     ServerConfigType     `koanf:"server" validate:"required"`
	Supervisor SupervisorConfigType `koanf:"supervisor" validate:"required"`
	Upstream   UpstreamConfigType   `koanf:"upstream" validate:"required"`
	LogLevel   string               `koanf:"log_level" validate:"required"`
}

// ...
var configSingleton *ConfigType
var configInitLock *sync.Once = &sync.Once{}

// ...
func LoadDotEnv() {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			logrus.Info("no .env file found")
		} else {
			logrus.WithError(err).Error("error reading .env file.")
		}
	} else {
		logrus.Info("found .env file")
	}
}

func doUnmarshal(cfg *ConfigType) error {
	k := koanf.New(".")
	if err := k.Load(confmap.Provider(getDefault(), "."), nil); err != nil {
		return err
	}
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "__", ".", -1)
	}), nil); err != nil {
		return err
	}
	return k.Unmarshal("", cfg)
}
func doValidate(cfg *ConfigType) error {
	validatr := validator.New(validator.WithRequiredStructEnabled())
	return validatr.Struct(cfg)
}
func getDefault() map[string]interface{} {
	return map[string]interface{}{
		"log_level":           "WARNING",
		"server.listen":       ":3000",
		"supervisor.interval": 60,
	}
}

func Config() *ConfigType {
	configInitLock.Do(func() {
		configSingleton = &ConfigType{}
		if err := doUnmarshal(configSingleton); err != nil {
			logrus.Fatalf("error parsing config: %s", err)
		}
		if err := doValidate(configSingleton); err != nil {
			logrus.Fatalf("error validating config: %s", err)
		}
		logrus.WithField("cfg", fmt.Sprintf("%+v", configSingleton)).Warn("config loaded")
	})
	return configSingleton
}
