package supervisor

import (
	"fmt"

	"github.com/anonopiran/Fly2User/internal/v2ray"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserRecord struct {
	// gorm.Model
	ID    uint
	UUID  string `gorm:"unique"`
	Email string `gorm:"unique"`
	Level uint32
}

func (rec *UserRecord) logger(ll *logrus.Entry) *logrus.Entry {
	if ll == nil {
		ll = logrus.NewEntry(logrus.StandardLogger())
	}
	return ll.WithField("user", fmt.Sprintf("%+v", rec))
}
func (rec *UserRecord) AsV2ray() *v2ray.UserType {
	return &v2ray.UserType{
		Email:  rec.Email,
		Secret: rec.UUID,
		Level:  rec.Level,
	}
}

// ...

func NewUserORM(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("error connecting user db: %s", err)
	}
	_d, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("can not get connection poll %s", err)
	}
	_d.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&UserRecord{}); err != nil {
		return nil, fmt.Errorf("error migrating user db: %s", err)
	}
	return db, nil
}
