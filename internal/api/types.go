package api

import "github.com/anonopiran/Fly2User/internal/supervisor"

type AddUserReq struct {
	UUID  string `json:"uuid" binding:"required"`
	Email string `gorm:"email" binding:"required"`
	Level uint32 `gorm:"level"`
}
type RmUserReq struct {
	Email string `gorm:"email" binding:"required"`
}

func (r *AddUserReq) AsUSer() *supervisor.UserRecord {
	return &supervisor.UserRecord{
		UUID:  r.UUID,
		Email: r.Email,
		Level: r.Level,
	}
}
func (r *RmUserReq) AsUSer() *supervisor.UserRecord {
	return &supervisor.UserRecord{
		Email: r.Email,
	}
}
