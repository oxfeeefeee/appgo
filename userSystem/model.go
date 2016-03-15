package userSystem

import (
	"github.com/oxfeeefeee/appgo"
	"time"
)

type UserModel struct {
	Id           appgo.Id
	CreatedAt    time.Time
	DeletedAt    time.Time
	Username     string `gorm:"size:63;unique_index"`
	PasswordSalt []byte `gorm:"size:63"`
	PasswordHash []byte `gorm:"size:63"`
	Role         appgo.Role
	// Only store appToken because webTokens are short lived
	AppToken    []byte `gorm:"size:63"`
	Email       string `gorm:"unique_index"`
	BannedUntil time.Time

	WeiboId       string `gorm:"size:63;unique_index"`
	WeixinUnionId string `gorm:"size:63;unique_index"`

	Nickname string `gorm:"size:63"`
	Portrait string `gorm:"size:255"`
	Sex      appgo.Sex
}

func (_ *UserModel) TableName() string {
	return "users"
}
