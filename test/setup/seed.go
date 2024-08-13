package test_setup

import (
	"github.com/anonopiran/Fly2User/internal/v2ray"
	"github.com/google/uuid"
)

func NewUserSeed() UserSeed {
	d := UserSeed{
		Usrs: []v2ray.UserType{{
			Email:  "test@user.com",
			Secret: uuid.NewString(),
			Level:  0,
		}, {
			Email:  "test@user2.com",
			Secret: uuid.NewString(),
			Level:  0,
		},
		},
	}
	d.DupUserEmail = v2ray.UserType{
		Email:  d.Usrs[0].Email,
		Secret: uuid.NewString(),
		Level:  0,
	}
	d.DupUserUUID = v2ray.UserType{
		Email:  "test@user3.com",
		Secret: d.Usrs[0].Secret,
		Level:  0,
	}
	return d
}

type UserSeed struct {
	Usrs         []v2ray.UserType
	DupUserEmail v2ray.UserType
	DupUserUUID  v2ray.UserType
}
