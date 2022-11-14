package config

import (
	u2g "Fly2User/User2Grpc"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var lock = &sync.Mutex{}

// ==================================
// types
// ==================================
type Path string
type LogLevel string
type InboundListType []u2g.InboundType
type settings_type struct {
	LogLevel          LogLevel        `env:"LOG_LEVEL" env-default:"warning"`
	V2flyApiAddress   string          `env:"V2FLY_API_ADDRESS" env-required:"true"`
	SuperviseInterval int             `env:"SUPERVISE_INTERVAL" env-default:"5"`
	SuperviseUuid     string          `env:"SUPERVISE_UUID" env-required:"true"`
	UserDir           Path            `env:"USER_DIR" env-default:"./storage/users"`
	InboundList       InboundListType `env:"INBOUND_LIST" env-required:"true"`
}

var conf *settings_type

// ==================================
// functions
// ==================================
func Config() *settings_type {
	if conf == nil {
		lock.Lock()
		defer lock.Unlock()
		if conf == nil {
			load_dot_env()
			var _conf settings_type
			err := cleanenv.ReadEnv(&_conf)
			if err != nil {
				log.WithError(err).Panic("can not initiate configuration")
			}
			conf = &_conf
			log.WithField("data", fmt.Sprintf("%+v", conf)).Debug("Parsed Configuration")
		}
	}
	return conf
}
func Describe() {
	var cfg settings_type
	help, err := cleanenv.GetDescription(&cfg, nil)
	if err != nil {
		log.WithError(err).Panic("can not generate description")
	}
	log.Println(help)
}
func load_dot_env() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Info("no .env file found")
	}
}

func (f *Path) SetValue(s string) error {
	err := os.MkdirAll(s, os.ModePerm)
	if err != nil {
		log.WithError(err).Error("can not create dir")
		return err
	}
	*f = Path(s)
	return nil
}
func (f *LogLevel) SetValue(s string) error {
	ll, err := log.ParseLevel(s)
	if err != nil {
		log.WithError(err).Error("can not set log level")
		return err
	}
	log.SetLevel(ll)
	*f = LogLevel(s)
	return nil
}
func (f *InboundListType) SetValue(s string) error {
	inbound_list := InboundListType{}
	for _, i_ := range strings.Split(s, ",") {
		id := strings.Split(i_, ":")
		ita := id[0]
		ity, err := str2inbtype(id[1])
		if err != nil {
			log.WithError(err).Error("inbound type not understood")
		}
		inbound_list = append(inbound_list, u2g.InboundType{Tag: ita, Proto: ity})
	}
	if len(inbound_list) == 0 {
		return errors.New("no inbound found")
	}
	*f = inbound_list
	return nil
}
func str2inbtype(s string) (u2g.PROTO, error) {
	v := new(u2g.PROTO)
	switch strings.ToUpper(s) {
	case "VMESS":
		*v = u2g.VMESS_PROTO
	case "VLESS":
		*v = u2g.VLESS_PROTO
	case "TROJAN":
		*v = u2g.TROJAN_PROTO
	default:
		return u2g.NULL_PROTO, fmt.Errorf("proto %s not found", s)
	}
	return *v, nil
}
