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
	"github.com/oxfeeefeee/appgo/services/qiniu"
	"github.com/oxfeeefeee/appgo/services/qq"
	"github.com/oxfeeefeee/appgo/services/weibo"
	"github.com/oxfeeefeee/appgo/services/weixin"
	"github.com/oxfeeefeee/appgo/toolkit/crypto"
	"github.com/parnurzeal/gorequest"
)

var U *UserSystem

type OnCreatedCallback func(id appgo.Id) error

type UserDataFromOAuthCode func(code string) (*UserData, string, error)

type UserSystem struct {
	db            *gorm.DB
	tableName     string
	Pushers       map[string]appgo.Pusher
	DefaultPusher appgo.Pusher
	OnCreated     OnCreatedCallback
	OAuths        []UserDataFromOAuthCode
	appgo.MobileMsgSender
	appgo.KvStore
}

type UserSystemSettings struct {
	TableName       string
	Pushers         []appgo.Pusher
	MobileMsgSender appgo.MobileMsgSender
	KvStore         appgo.KvStore
	OnCreated       OnCreatedCallback
	OAuths          []UserDataFromOAuthCode
}

type UserData struct {
	Id       appgo.Id
	Username *string
	Email    *string
	Mobile   *string
	Role     appgo.Role
	Platform appgo.Platform
	Nickname *string
	Portrait *string
	Sex      appgo.Sex
}

func Init(db *gorm.DB, settings UserSystemSettings) *UserSystem {
	if U != nil {
		return U
	}
	sender := settings.MobileMsgSender
	if sender == nil {
		sender = &defaultSender{}
	}
	store := settings.KvStore
	if store == nil {
		store = &defaultKvStore{}
	}
	pushers := make(map[string]appgo.Pusher)
	for _, p := range settings.Pushers {
		if _, ok := pushers[p.Name()]; ok {
			panic("Duplicated pusher names.")
		}
		pushers[p.Name()] = p
	}
	U = &UserSystem{
		db,
		settings.TableName,
		pushers,
		&defaultPusher{},
		settings.OnCreated,
		settings.OAuths,
		sender,
		store,
	}
	initTable(db)
	mobileSupport := U
	if settings.MobileMsgSender == nil || settings.KvStore == nil {
		mobileSupport = nil
	}
	auth.Init(U, U, U, U, mobileSupport, U)
	return U
}

func (u *UserSystem) Validate(token auth.Token) bool {
	return true //todo
}

func (u *UserSystem) GetUserModel(id appgo.Id) (*UserModel, error) {
	user := &UserModel{Id: id}
	if err := u.db.First(user).Error; err != nil {
		log.WithFields(log.Fields{
			"id":        id,
			"gormError": err,
		}).Infoln("failed to find user")
		return nil, appgo.NotFoundErr
	}
	return user, nil
}

func (u *UserSystem) IsBanned(id appgo.Id) bool {
	user := &UserModel{Id: id}
	if err := u.db.Select("banned_until").First(user).Error; err != nil {
		return false
	} else {
		return user.BannedUntil != nil
	}
}

func (u *UserSystem) RecordLastActiveAt(uid appgo.Id) error {
	user := &UserModel{Id: uid}
	return u.db.Model(user).Update("last_active_at", gorm.Expr("now()")).Error
}

func (u *UserSystem) CheckIn(id appgo.Id, role appgo.Role,
	newToken auth.Token) (banned bool, extraInfo interface{}, err error) {
	user, err := u.GetUserModel(id)
	if err != nil {
		return false, nil, err
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

func (u *UserSystem) UpdatePushInfo(id appgo.Id, pushInfo *appgo.PushInfo) error {
	if pushInfo == nil {
		return nil
	}
	user := &UserModel{Id: id}
	var updates UserModel
	updates.Platform = pushInfo.Platform
	updates.Manufacturer = sql.NullString{pushInfo.Manufacturer, true}
	updates.PushProvider = sql.NullString{pushInfo.Provider, true}
	updates.PushToken = sql.NullString{pushInfo.Token, true}
	if err := u.db.Model(user).Updates(&updates).Error; err != nil {
		log.WithFields(log.Fields{
			"id":        id,
			"gormError": err,
		}).Errorln("failed to save push info")
		return appgo.InternalErr
	}
	return nil
}

func (u *UserSystem) GetWeixinUser(unionId string) (appgo.Id, error) {
	return getUser(u.db, &UserModel{
		WeixinUnionId: database.SqlStr(unionId)})
}

func (u *UserSystem) AddWeixinUser(info *weixin.UserInfo) (appgo.Id, error) {
	sex := appgo.SexDefault
	if info.Sex == 1 {
		sex = appgo.SexMale
	} else if info.Sex == 2 {
		sex = appgo.SexFemale
	}
	imageName := "weixin_user_" + info.UnionId
	imageName, _ = copyImage(info.Image, imageName)
	user := &UserModel{
		Role:          appgo.RoleAppUser,
		WeixinUnionId: database.SqlStr(info.UnionId),
		Nickname:      database.SqlStr(info.Nickname),
		Portrait:      database.SqlStr(imageName),
		Sex:           sex,
	}
	return u.saveUser(user)
}

func (u *UserSystem) GetWeiboUser(openId string) (appgo.Id, error) {
	return getUser(u.db, &UserModel{
		WeiboId: database.SqlStr(openId)})
}

func (u *UserSystem) AddWeiboUser(info *weibo.UserInfo) (appgo.Id, error) {
	sex := appgo.SexDefault
	if info.Sex == "m" {
		sex = appgo.SexMale
	} else if info.Sex == "f" {
		sex = appgo.SexFemale
	}
	imageName := "weibo_user_" + info.Id
	imageName, _ = copyImage(info.Image, imageName)
	user := &UserModel{
		Role:     appgo.RoleAppUser,
		WeiboId:  database.SqlStr(info.Id),
		Nickname: database.SqlStr(info.Name),
		Portrait: database.SqlStr(imageName),
		Sex:      sex,
	}
	return u.saveUser(user)
}

func (u *UserSystem) GetQqUser(openId string) (appgo.Id, error) {
	return getUser(u.db, &UserModel{
		QqOpenId: database.SqlStr(openId)})
}

func (u *UserSystem) AddQqUser(info *qq.UserInfo) (appgo.Id, error) {
	sex := appgo.SexDefault
	if info.Sex == "男" {
		sex = appgo.SexMale
	} else if info.Sex == "女" {
		sex = appgo.SexFemale
	}
	imageName := "qq_user_" + info.OpenId
	imageName, _ = copyImage(info.Image, imageName)
	user := &UserModel{
		Role:     appgo.RoleAppUser,
		QqOpenId: database.SqlStr(info.OpenId),
		Nickname: database.SqlStr(info.Nickname),
		Portrait: database.SqlStr(imageName),
		Sex:      sex,
	}
	return u.saveUser(user)
}

func (u *UserSystem) GetOAuthUser(index int, id string) (appgo.Id, error) {
	if index < 0 || index > 1 {
		return 0, errors.New("Invalid index")
	}
	if index == 0 {
		return getUser(u.db, &UserModel{
			OAuth0Id: database.SqlStr(id)})
	} else {
		return getUser(u.db, &UserModel{
			OAuth1Id: database.SqlStr(id)})
	}
}

func (u *UserSystem) AddOAuthUser(index int, id string, ui interface{}) (appgo.Id, error) {
	if index < 0 || index > 1 {
		return 0, errors.New("Invalid index")
	}
	userInfo, ok := ui.(*UserData)
	if !ok {
		return 0, errors.New("Invalid user data")
	}
	var user UserModel
	if index == 0 {
		user.OAuth0Id = database.SqlStr(id)
	} else {
		user.OAuth1Id = database.SqlStr(id)
	}
	user.Role = appgo.RoleAppUser
	user.Nickname = database.SqlStr(*userInfo.Nickname)
	user.Portrait = database.SqlStr(*userInfo.Portrait)
	user.Sex = userInfo.Sex
	return u.saveUser(&user)
}

func (u *UserSystem) GetOAuthUserInfo(index int, code string) (interface{}, string, error) {
	if index < 0 || index > 1 || index >= len(u.OAuths) {
		return nil, "", errors.New("Invalid index")
	}
	return u.OAuths[index](code)
}

func (u *UserSystem) HasMobileUser(mobile string) (bool, error) {
	where := &UserModel{Mobile: database.SqlStr(mobile)}
	var user UserModel
	if db := u.db.Where(where).First(&user); db.Error != nil {
		if db.RecordNotFound() {
			return false, nil
		}
		return false, db.Error
	}
	return true, nil
}

func (u *UserSystem) GetMobileUser(mobile, password string) (appgo.Id, error) {
	where := &UserModel{Mobile: database.SqlStr(mobile)}
	var user UserModel
	if db := u.db.Where(where).First(&user); db.Error != nil {
		if db.RecordNotFound() {
			return 0, nil
		}
		return 0, db.Error
	}
	if len(user.PasswordSalt) == 0 || len(user.PasswordHash) == 0 {
		return 0, errors.New("password not set yet")
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
			Portrait:     database.SqlStr(info.Portrait),
			Sex:          info.Sex,
		}
		return u.saveUser(user)
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

func (u *UserSystem) SetMobileForUser(mobile string, userId appgo.Id) error {
	user := &UserModel{Id: userId}
	var updates UserModel
	updates.Mobile = sql.NullString{mobile, true}
	if err := u.db.Model(user).Updates(&updates).Error; err != nil {
		log.WithFields(log.Fields{
			"id":        userId,
			"gormError": err,
		}).Errorln("failed to update mobile")
		return appgo.InternalErr
	}
	return nil
}

func (u *UserSystem) UpdateAppToken(id appgo.Id, role appgo.Role) (string, error) {
	newToken := auth.NewToken(id, role)
	tk := sql.NullString{string(newToken), true}
	user := &UserModel{Id: id}
	if err := u.db.Model(user).Updates(&UserModel{AppToken: tk}).Error; err != nil {
		log.WithFields(log.Fields{
			"id":        id,
			"gormError": err,
		}).Errorln("failed to save token")
		return "", appgo.InternalErr
	}
	return tk.String, nil
}

func (u *UserSystem) saveUser(user *UserModel) (appgo.Id, error) {
	if db := u.db.Save(user); db.Error != nil {
		return 0, db.Error
	}
	if u.OnCreated != nil {
		return user.Id, u.OnCreated(user.Id)
	}
	return user.Id, nil
}

func dbModelToData(model *UserModel) *UserData {
	return &UserData{
		Id:       model.Id,
		Username: &model.Username.String,
		Email:    &model.Email.String,
		Mobile:   &model.Mobile.String,
		Role:     model.Role,
		Platform: model.Platform,
		Nickname: &model.Nickname.String,
		Portrait: &model.Portrait.String,
		Sex:      model.Sex,
	}
}

func dataToDbModel(data *UserData) *UserModel {
	var um UserModel
	um.Id = data.Id
	if data.Username != nil {
		um.Username.String, um.Username.Valid = *data.Username, true
	}
	if data.Email != nil {
		um.Email.String, um.Email.Valid = *data.Email, true
	}
	if data.Mobile != nil {
		um.Mobile.String, um.Mobile.Valid = *data.Mobile, true
	}
	um.Role = data.Role
	if data.Nickname != nil {
		um.Nickname.String, um.Nickname.Valid = *data.Nickname, true
	}
	if data.Portrait != nil {
		um.Portrait.String, um.Portrait.Valid = *data.Portrait, true
	}
	um.Sex = data.Sex
	return &um
}

func copyImage(from, to string) (string, error) {
	_, body, errs := gorequest.New().Get(from).End()
	if errs != nil {
		errstr := "Failed to get image: " + from
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    from,
		}).Error(errstr)
		return "", errors.New(errstr)
	}
	return qiniu.PutFile(to, bytes.NewReader([]byte(body)), int64(len(body)))
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
	} else {
		db.AutoMigrate(&user)
	}
}
