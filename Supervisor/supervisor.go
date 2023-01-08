package supervisor

import (
	config "Fly2User/Config"
	u2g "Fly2User/User2Grpc"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

var cfg = config.Config()

func AddAllUsers() (int, error) {
	files, err := os.ReadDir(string(cfg.UserDir))
	if err != nil {
		log.WithError(err).Error("can not read user files")
		return 0, err
	}
	conn, err := u2g.NewGrpcConn(cfg.V2flyApiAddress)
	if err != nil {
		log.WithError(err).Error("error while dialing grpc server")
		return 0, err
	}
	defer conn.Close()
	cnt_total := 0
	for _, f_ := range files {
		email := f_.Name()
		email = email[:len(email)-len(".user")]
		log_w_udata := log.WithField("email", email)
		// ....
		if err != nil {
			log_w_udata.WithError(err).Error("can not parse json file")
			continue
		}
		f_data, err := UserFromFile(email)
		if err != nil {
			continue
		}
		add_err := f_data.AddMultiple(conn, &cfg.InboundList, false)
		if len(add_err) > 0 {
			for _, e_ := range add_err {
				log_w_udata.WithField("inbound", e_.GetInbound()).WithError(e_)
			}
		}
		cnt_total += 1
	}
	log.Infof("%d users added", cnt_total)
	return cnt_total, nil
}
func Supervise() {
	sleeper := time.NewTicker(time.Second * time.Duration(cfg.SuperviseInterval))
	for {
		if CheckServerRestart() {
			log.Warning("found server restart. adding users")
			AddAllUsers()
		} else {
			log.Debug("no server restart found. skipping")
		}
		<-sleeper.C
	}
}
func UserFromFile(email string) (u2g.UserAddType, error) {
	f_path := user_fpath(email)
	log_w_data := log.WithField("email", email)
	f_file, err := os.ReadFile(f_path)
	f_data := u2g.UserAddType{}
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
func UserToFile(user u2g.UserAddType) error {
	email := user.Email
	f_path := user_fpath(email)
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
	f_path := user_fpath(email)
	log_w_data := log.WithField("email", email)
	if err := os.Remove(f_path); err != nil && !os.IsNotExist(err) {
		log_w_data.Error(err)
		return err
	}
	return nil
}
func UserCheckFile(email string) bool {
	f_path := user_fpath(email)
	_, err := os.Stat(f_path)
	return err == nil
}
func CheckServerRestart() bool {
	conn, err := u2g.NewGrpcConn(cfg.V2flyApiAddress)
	if err != nil {
		log.WithError(err).Error("error while watching server restart")
		return false
	}
	defer conn.Close()
	_u := u2g.UserAddType{
		Uuid:  cfg.SuperviseUuid,
		Email: cfg.SuperviseUuid + "@supervis.or",
		Level: 0,
	}
	err = _u.Add(conn, &cfg.InboundList[0], true)
	return err == nil || !err.IsUserExistsError()
}
func user_fpath(email string) string {
	return filepath.Join(string(cfg.UserDir), email+".user")
}
