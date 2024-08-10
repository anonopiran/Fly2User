package supervisor

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type downServer struct {
	// gorm.Model
	ID        uint
	UpSrvId   uint
	IpAddress string `gorm:"unique"`
}

func (ds *downServer) logger(ll *logrus.Entry) *logrus.Entry {
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
	if err := db.AutoMigrate(&downServer{}); err != nil {
		return nil, fmt.Errorf("error migrating downserver db: %s", err)
	}
	return db, nil
}
func newDownServer(upSrvId uint, ipAddress string) *downServer {
	return &downServer{
		UpSrvId:   upSrvId,
		IpAddress: ipAddress,
	}
}
