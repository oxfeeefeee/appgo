package userSystem

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/oxfeeefeee/appgo"
	"github.com/oxfeeefeee/appgo/auth"
	"github.com/oxfeeefeee/appgo/services/weibo"
	"github.com/oxfeeefeee/appgo/services/weixin"
)

var U *UserSystem

type UserSystem struct {
	db *gorm.DB
}

func Init(db *gorm.DB) *UserSystem {
	if U != nil {
		return U
	}
	var user UserModel
	if !db.HasTable(&user) {
		db.CreateTable(&user)
	}
	U = &UserSystem{
		db,
	}
	auth.Init(U, U, U)
	return U
}

func (u *UserSystem) Validate(token auth.Token) bool {
	return true //todo
}

func (u *UserSystem) CheckIn(id appgo.Id, role appgo.Role,
	newToken auth.Token) (banned bool, extraInfo interface{}, err error) {
	var user UserModel
	if err := u.db.First(&user, id).Error; err != nil {
		log.WithFields(log.Fields{
			"id":        id,
			"gormError": err,
		}).Infoln("user not found")
		return false, nil, appgo.NotFoundErr
	}
	if role > user.Role {
		return false, nil, appgo.ForbiddenErr
	}
	// TODO save other tokens
	if newToken != "" && role == appgo.RoleAppUser {
		if err := u.db.Save(&user).Error; err != nil {
			log.WithFields(log.Fields{
				"id":        id,
				"gormError": err,
			}).Errorln("failed to save token")
		}
		return false, nil, appgo.InternalErr
	}
	// todo stuff about ban
	return false, &user, nil
}

func (u *UserSystem) GetWeixinUser(unionId string) (uid appgo.Id, err error) {
	var user UserModel
	err = u.db.Where(&UserModel{WeixinUnionId: unionId}).First(&user).Error
	if err != nil {
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
		WeixinUnionId: info.UnionId,
		Nickname:      info.Nickname,
		Portrait:      info.Image, //todo: copy image
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
	err = u.db.Where(&UserModel{WeiboId: openId}).First(&user).Error
	if err != nil {
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
		WeiboId:  info.Id,
		Nickname: info.Name,
		Portrait: info.Image, //todo: copy image
		Sex:      sex,
	}
	db := u.db.Save(user)
	if db.Error != nil {
		return
	}
	return user.Id, nil
}
