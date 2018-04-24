package userSystem

import (
	"database/sql"
	"github.com/oxfeeefeee/appgo"
	"time"
)

type UserModel struct {
	Id            appgo.Id
	Username      sql.NullString `gorm:"size:63;unique_index"`
	Email         sql.NullString `gorm:"size:63;unique_index"`
	Mobile        sql.NullString `gorm:"size:15;unique_index"`
	PasswordSalt  []byte         `gorm:"size:64"`
	PasswordHash  []byte         `gorm:"size:64"`
	WeiboId       sql.NullString `gorm:"size:63;unique_index"`
	WeixinUnionId sql.NullString `gorm:"size:63;unique_index"`
	QqOpenId      sql.NullString `gorm:"size:63;unique_index"`
	OAuth0Id      sql.NullString `gorm:"size:63;unique_index"`
	OAuth1Id      sql.NullString `gorm:"size:63;unique_index"`
	// Only store appToken because webTokens are short lived
	AppToken     sql.NullString `gorm:"size:127"`
	Role         appgo.Role
	Platform     appgo.Platform
	Manufacturer sql.NullString `gorm:"size:127"`
	PushProvider sql.NullString `gorm:"size:16"`
	PushToken    sql.NullString `gorm:"size:127"`
	Nickname     sql.NullString `gorm:"size:63"`
	RealName     sql.NullString `gorm:"size:63"`
	Portrait     sql.NullString `gorm:"size:255"`
	Sex          appgo.Sex
	BannedUntil  *time.Time
	CreatedAt    time.Time
	LastActiveAt time.Time `gorm:"index"`
	DeletedAt    *time.Time
}

func (_ *UserModel) TableName() string {
	if U.tableName != "" {
		return U.tableName
	}
	return "users"
}
