package userSystem

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/oxfeeefeee/appgo/database"
	"github.com/oxfeeefeee/appgo/services/weibo"
	"github.com/oxfeeefeee/appgo/services/weixin"
)

var U *UserSystem

type UserSystem struct {
	db        *gorm.DB
	tableName string
}

func Init(db *gorm.DB, tableName string) *UserSystem {
	if U != nil {
		return U
	}
	U = &UserSystem{
		db,
		tableName,
	}
	initTable(db)
	auth.Init(U, U, U)
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
	var user UserModel
	err = u.db.Where(&UserModel{
		WeixinUnionId: database.SqlStr(unionId)}).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}
	return user.Id, nil
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
	db := u.db.Save(user)
	if db.Error != nil {
		return
	}
	return user.Id, nil
}

func (u *UserSystem) GetWeiboUser(openId string) (uid appgo.Id, err error) {
	var user UserModel
	err = u.db.Where(&UserModel{
		WeiboId: database.SqlStr(openId)}).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}
	return user.Id, nil
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
	db := u.db.Save(user)
	if db.Error != nil {
		return
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
