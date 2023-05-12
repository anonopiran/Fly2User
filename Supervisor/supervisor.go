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

// var servers []config.V2rayUrlType
// var inbounds []u2u.InboundType
// var userDir config.PathType
// var supervisUUID string

// ==================================
// functions
// ==================================
func Supervise() {
	sleeper := time.NewTicker(time.Second * time.Duration(cfg().SuperviseInterval))
	for {
		for _, trg := range cfg().V2flyApiAddress {
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
					users, err := ReadAllUsers()
					if err != nil {
						log.WithError(err).Error("can not read user files")
						continue
					}
					AddManyUsers(&users, &srv)
				} else {
					log_w_srv.Debug("no server restart found. skipping")
				}
			}
		}
		<-sleeper.C
	}
}
func ReadAllUsers() ([]u2u.UserAddType, error) {
	var add []u2u.UserAddType
	files, err := os.ReadDir(string(cfg().UserDir))
	if err != nil {
		return add, err
	}
	for _, f_ := range files {
		if !strings.HasSuffix(f_.Name(), ".user") {
			continue
		}
		email := f_.Name()
		email = email[:len(email)-len(".user")]
		f_data, err := UserFromFile(email)
		if err != nil {
			continue
		}
		add = append(add, f_data)
	}
	return add, nil
}
func AddManyUsers(users *[]u2u.UserAddType, target *u2u.ServerType) (int, error) {
	conn, err := target.DialGrpc()
	if err != nil {
		log.WithError(err).Error("error while dialing grpc server")
		return 0, err
	}
	defer conn.Close()
	cnt_total := 0
	for _, f_ := range *users {
		i_list := cfg().InboundList
		log_w_udata := log.WithField("email", f_.Email)
		add_err := f_.AddMultiple(conn, &i_list, false)
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
	i_list := cfg().InboundList
	_u := u2u.UserAddType{
		Uuid:  string(cfg().SuperviseUuid),
		Email: string(cfg().SuperviseUuid) + "@supervis.or",
		Level: 0,
	}
	err_create := _u.AddMultiple(conn, &i_list, true)
	if len(err_create) > 0 {
		for _, err := range err_create {
			if !err.IsUserExistsError() {
				log.WithError(err).Panic("some error happened while add supervisor user")
			}
		}
	}
	return len(err_create) != len(cfg().InboundList)
}
func UserFilePath(email string) string {
	return filepath.Join(string(cfg().UserDir), email+".user")
}

func cfg() config.SettingsType {
	return *config.Config()
}
