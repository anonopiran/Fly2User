package supervisor

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DownServerType struct {
	// gorm.Model
	ID        uint
	UpSrvId   uint
	IpAddress string `gorm:"unique"`
}

func (ds *DownServerType) logger(ll *logrus.Entry) *logrus.Entry {
	if ll == nil {
		ll = logrus.NewEntry(logrus.StandardLogger())
	}
	return ll.WithField("server", ds.IpAddress)
}

func newDownServerDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	if err != nil {
		return nil, fmt.Errorf("error creating downserver db: %s", err)
	}
	_d, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("can not get connection poll %s", err)
	}
	_d.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&DownServerType{}); err != nil {
		return nil, fmt.Errorf("error migrating downserver db: %s", err)
	}
	return db, nil
}
func newDownServer(upSrvId uint, ipAddress string) *DownServerType {
	return &DownServerType{
		UpSrvId:   upSrvId,
		IpAddress: ipAddress,
	}
}
