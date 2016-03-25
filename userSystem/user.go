package userSystem

import (
	"bytes"
	"database/sql"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/oxfeeefeee/appgo/database"
	"github.com/oxfeeefeee/appgo/services/qq"
	"github.com/oxfeeefeee/appgo/services/weibo"
	"github.com/oxfeeefeee/appgo/services/weixin"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
)

var U *UserSystem

type UserSystem struct {
	db        *gorm.DB
	tableName string
	appgo.MobileMsgSender
	appgo.KvStore
}

type UserSystemSettings struct {
	TableName       string
	MobileMsgSender appgo.MobileMsgSender
	KvStore         appgo.KvStore
}

func Init(db *gorm.DB, settings UserSystemSettings) *UserSystem {
	if U != nil {
		return U
	}
	U = &UserSystem{
		db,
		settings.TableName,
		settings.MobileMsgSender,
		settings.KvStore,
	}
	initTable(db)
	mobileSupport := U
	if settings.MobileMsgSender == nil || settings.KvStore == nil {
		mobileSupport = nil
	}
	auth.Init(U, U, U, U, mobileSupport)
	return U
}

func (u *UserSystem) Validate(token auth.Token) bool {
	return true //todo
}

func (u *UserSystem) CheckIn(id appgo.Id, role appgo.Role,
	newToken auth.Token) (banned bool, extraInfo interface{}, err error) {
	user := &UserModel{Id: id}
	if err := u.db.First(user).Error; err != nil {
		log.WithFields(log.Fields{
			"id":        id,
			"gormError": err,
		}).Infoln("failed to find user")
		return false, nil, appgo.NotFoundErr
	}
	if role > user.Role {
		return false, nil, appgo.ForbiddenErr
	}
	// TODO save other tokens
	if newToken != "" && role == appgo.RoleAppUser {
		tk := sql.NullString{string(newToken), true}
		if err := u.db.Model(user).Updates(&UserModel{AppToken: tk}).Error; err != nil {
			log.WithFields(log.Fields{
				"id":        id,
				"gormError": err,
			}).Errorln("failed to save token")
			return false, nil, appgo.InternalErr
		}
	}
	// todo stuff about ban
	return false, user, nil
}

func (u *UserSystem) GetWeixinUser(unionId string) (uid appgo.Id, err error) {
	return getUser(u.db, &UserModel{
		WeixinUnionId: database.SqlStr(unionId)})
}

func (u *UserSystem) AddWeixinUser(info *weixin.UserInfo) (uid appgo.Id, err error) {
	sex := appgo.SexDefault
	if info.Sex == 1 {
		sex = appgo.SexMale
	} else if info.Sex == 2 {
		sex = appgo.SexFemale
	}
	user := &UserModel{
		Role:          appgo.RoleAppUser,
		WeixinUnionId: database.SqlStr(info.UnionId),
		Nickname:      database.SqlStr(info.Nickname),
		Portrait:      database.SqlStr(info.Image), //todo: copy image
		Sex:           sex,
	}
	return saveUser(u.db, user)
}

func (u *UserSystem) GetWeiboUser(openId string) (uid appgo.Id, err error) {
	return getUser(u.db, &UserModel{
		WeiboId: database.SqlStr(openId)})
}

func (u *UserSystem) AddWeiboUser(info *weibo.UserInfo) (uid appgo.Id, err error) {
	sex := appgo.SexDefault
	if info.Sex == "m" {
		sex = appgo.SexMale
	} else if info.Sex == "f" {
		sex = appgo.SexFemale
	}
	user := &UserModel{
		Role:     appgo.RoleAppUser,
		WeiboId:  database.SqlStr(info.Id),
		Nickname: database.SqlStr(info.Name),
		Portrait: database.SqlStr(info.Image), //todo: copy image
		Sex:      sex,
	}
	return saveUser(u.db, user)
}

func (u *UserSystem) GetQqUser(openId string) (uid appgo.Id, err error) {
	return getUser(u.db, &UserModel{
		QqOpenId: database.SqlStr(openId)})
}

func (u *UserSystem) AddQqUser(info *qq.UserInfo) (uid appgo.Id, err error) {
	sex := appgo.SexDefault
	if info.Sex == "男" {
		sex = appgo.SexMale
	} else if info.Sex == "女" {
		sex = appgo.SexFemale
	}
	user := &UserModel{
		Role:     appgo.RoleAppUser,
		QqOpenId: database.SqlStr(info.OpenId),
		Nickname: database.SqlStr(info.Nickname),
		Portrait: database.SqlStr(info.Image), //todo: copy image
		Sex:      sex,
	}
	return saveUser(u.db, user)
}

func (u *UserSystem) GetMobileUser(mobile, password string) (uid appgo.Id, err error) {
	where := &UserModel{Mobile: database.SqlStr(mobile)}
	var user UserModel
	if db := u.db.Where(where).First(&user); db.Error != nil {
		if db.RecordNotFound() {
			return 0, nil
		}
		return 0, db.Error
	}
	hash := crypto.SaltedHash(user.PasswordSalt, password)
	if !bytes.Equal(hash[:], user.PasswordHash) {
		return 0, appgo.InvalidPasswordErr
	}
	return user.Id, nil
}

func (u *UserSystem) AddMobileUser(info *auth.MobileUserInfo) (appgo.Id, error) {
	if salt, err := crypto.RandBytes(64); err != nil {
		return 0, err
	} else {
		hash := crypto.SaltedHash(salt, info.Password)
		user := &UserModel{
			Role:         appgo.RoleAppUser,
			Mobile:       database.SqlStr(info.Mobile),
			PasswordSalt: salt,
			PasswordHash: hash[:],
			Nickname:     database.SqlStr(info.Nickname),
			Sex:          info.Sex,
		}
		return saveUser(u.db, user)
	}
}

func (u *UserSystem) UpdatePwByMobile(mobile, password string) error {
	where := &UserModel{Mobile: database.SqlStr(mobile)}
	if uid, err := getUser(u.db, where); err != nil {
		return err
	} else if uid == 0 {
		return errors.New("UpdatePwByMobile: user not found")
	}
	if salt, err := crypto.RandBytes(64); err != nil {
		return err
	} else {
		hash := crypto.SaltedHash(salt, password)
		update := &UserModel{PasswordSalt: salt, PasswordHash: hash[:]}
		if err := u.db.Model(&UserModel{}).Where(where).
			Updates(update).Error; err != nil {
			return err
		}
		return nil
	}
}

func getUser(db *gorm.DB, where *UserModel) (appgo.Id, error) {
	var user UserModel
	if db := db.Where(where).First(&user); db.Error != nil {
		if db.RecordNotFound() {
			return 0, nil
		}
		return 0, db.Error
	}
	return user.Id, nil
}

func saveUser(db *gorm.DB, user *UserModel) (appgo.Id, error) {
	if db := db.Save(user); db.Error != nil {
		return 0, db.Error
	}
	return user.Id, nil
}

func initTable(db *gorm.DB) {
	var user UserModel
	if !db.HasTable(&user) {
		err := db.CreateTable(&user).Error
		if err != nil {
			log.WithFields(log.Fields{
				"gormError": err,
			}).Infoln("failed to create user table")
		}
		tableName := user.TableName()
		if err := db.Exec("ALTER TABLE " + tableName + " AUTO_INCREMENT=10001").Error; err != nil {
			if err != nil {
				log.WithFields(log.Fields{
					"gormError": err,
				}).Infoln("failed to alter user table AUTO_INCREMENT")
			}
		}
	}
}
