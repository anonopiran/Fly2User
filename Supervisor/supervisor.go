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
	conn, err := u2g.GrpcClient(cfg.V2flyApiAddress)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	cnt_total := 0
	for _, f_ := range files {
		email := f_.Name()
		email = email[:len(email)-len(".user")]
		log_w_data := log.WithField("email", email)
		// ....
		if err != nil {
			log_w_data.WithError(err).Error("can not parse json file")
			continue
		}
		f_data, err := UserFromFile(email)
		if err != nil {
			continue
		}
		add_err := u2g.AddUser(conn, cfg.InboundList, &f_data, false)
		if len(add_err) > 0 {
			log_w_data.WithError(err)
			continue
		}
		cnt_total += 1
	}
	log.Infof("%d users added", cnt_total)
	return cnt_total, nil
}
func Supervise() {
	sleeper := time.NewTicker(time.Second * time.Duration(cfg.SuperviseInterval))
	for {
		if check_server_restart() {
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
	err := os.Remove(f_path)
	if err != nil {
		log_w_data.Error(err)
		return err
	}
	return nil
}
func UserCheckFie(email string) bool {
	f_path := user_fpath(email)
	_, err := os.Stat(f_path)
	return err == nil
}
func user_fpath(email string) string {
	return filepath.Join(string(cfg.UserDir), email+".user")
}
func check_server_restart() bool {
	conn, err := u2g.GrpcClient(cfg.V2flyApiAddress)
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
	err_list := u2g.AddUser(conn, cfg.InboundList[:1], &_u, true)
	return !(len(err_list) > 0 && u2g.AddUserErrIsExists(err_list[0]))
}
