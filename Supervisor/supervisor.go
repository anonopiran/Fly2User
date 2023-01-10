package supervisor

import (
	config "Fly2User/Config"
	u2u "Fly2User/User2Upstream"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var servers []config.V2rayUrlType
var inbounds []u2u.InboundType
var userDir config.PathType
var supervisUUID string

// ==================================
// functions
// ==================================
func Supervise() {
	sleeper := time.NewTicker(time.Second * time.Duration(config.Config.SuperviseInterval))
	for {
		for _, trg := range servers {
			logWithTarget := log.WithField("target", trg)
			res, err := u2u.ResolveV2FlyServer(trg.AsUrl())
			if err != nil {
				logWithTarget.WithError(err).Error("error while resolving v2fly server")
				continue
			}
			for _, srv := range res {
				log_w_srv := log.WithField("server", srv)
				if CheckServerRestart(&srv) {
					log_w_srv.Warning("found server restart. adding users")
					AddAllUsers(&srv)
				} else {
					log_w_srv.Debug("no server restart found. skipping")
				}
			}
		}
		<-sleeper.C
	}
}
func AddAllUsers(target *u2u.ServerType) (int, error) {
	files, err := os.ReadDir(userDir.AsString())
	if err != nil {
		log.WithError(err).Error("can not read user files")
		return 0, err
	}
	conn, err := target.DialGrpc()
	if err != nil {
		log.WithError(err).Error("error while dialing grpc server")
		return 0, err
	}
	defer conn.Close()
	cnt_total := 0
	for _, f_ := range files {
		if !strings.HasSuffix(f_.Name(), ".user") {
			continue
		}
		email := f_.Name()
		email = email[:len(email)-len(".user")]
		log_w_udata := log.WithField("email", email)
		f_data, err := UserFromFile(email)
		if err != nil {
			continue
		}
		add_err := f_data.AddMultiple(conn, &inbounds, false)
		if len(add_err) > 0 {
			for _, e_ := range add_err {
				log_w_udata.WithField("inbound", e_.GetInbound()).WithError(e_).Error("")
			}
			continue
		}
		cnt_total += 1
	}
	log.Infof("%d users added", cnt_total)
	return cnt_total, nil
}
func UserFromFile(email string) (u2u.UserAddType, error) {
	f_path := UserFilePath(email)
	log_w_data := log.WithField("email", email)
	f_file, err := os.ReadFile(f_path)
	f_data := u2u.UserAddType{}
	if err != nil {
		log_w_data.WithError(err).Error("can not read json file")
		return f_data, err
	}
	err = json.Unmarshal(f_file, &f_data)
	if err != nil {
		log_w_data.WithError(err).Error("can not parse json file")
		return f_data, err
	}
	return f_data, nil
}
func UserToFile(user u2u.UserAddType) error {
	email := user.Email
	f_path := UserFilePath(email)
	log_w_data := log.WithField("email", email)
	jdata, err := json.Marshal(user)
	if err != nil {
		log_w_data.WithError(err).Error("can not marshal user to json")
		return err
	}
	os.WriteFile(f_path, jdata, 0644)
	return nil
}
func UserDelFile(email string) error {
	f_path := UserFilePath(email)
	log_w_data := log.WithField("email", email)
	if err := os.Remove(f_path); err != nil && !os.IsNotExist(err) {
		log_w_data.Error(err)
		return err
	}
	return nil
}
func UserCheckFile(email string) bool {
	f_path := UserFilePath(email)
	_, err := os.Stat(f_path)
	return err == nil || !os.IsNotExist(err)
}
func CheckServerRestart(target *u2u.ServerType) bool {
	conn, err := target.DialGrpc()
	if err != nil {
		log.WithError(err).Error("error while watching server restart")
		return false
	}
	defer conn.Close()
	_u := u2u.UserAddType{
		Uuid:  supervisUUID,
		Email: supervisUUID + "@supervis.or",
		Level: 0,
	}
	err_create := _u.AddMultiple(conn, &inbounds, true)
	if len(err_create) > 0 {
		for _, err := range err_create {
			if !err.IsUserExistsError() {
				log.WithError(err).Panic("some error happened while add supervisor user")
			}
		}
	}
	return len(err_create) != len(inbounds)
}
func UserFilePath(email string) string {
	return filepath.Join(string(userDir), email+".user")
}

func init() {
	servers = config.Config.V2flyApiAddress
	inbounds = config.Config.InboundList
	userDir = config.Config.UserDir
	supervisUUID = config.Config.SuperviseUuid.AsString()
}
