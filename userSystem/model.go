package userSystem

import (
	"database/sql"
	"github.com/oxfeeefeee/appgo"
	"time"
)

type UserModel struct {
	Id            appgo.Id
	Username      sql.NullString `gorm:"size:63;unique_index"`
	Email         sql.NullString `gorm:"unique_index"`
	PasswordSalt  []byte         `gorm:"size:63"`
	PasswordHash  []byte         `gorm:"size:63"`
	WeiboId       sql.NullString `gorm:"size:63;unique_index"`
	WeixinUnionId sql.NullString `gorm:"size:63;unique_index"`
	// Only store appToken because webTokens are short lived
	AppToken    sql.NullString `gorm:"size:127"`
	Role        appgo.Role
	Nickname    sql.NullString `gorm:"size:63"`
	Portrait    sql.NullString `gorm:"size:255"`
	Sex         appgo.Sex
	BannedUntil *time.Time
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

func (_ *UserModel) TableName() string {
	if U.tableName != "" {
		return U.tableName
	}
	return "users"
}
