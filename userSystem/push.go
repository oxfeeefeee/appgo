package userSystem

import (
	"database/sql"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
)

func (u *UserSystem) PushTo(users []appgo.Id, content *appgo.PushData) {
	tokens, err := u.GetPushTokens(users)
	if err != nil {
		log.WithFields(log.Fields{
			"users": users,
			"error": err,
		}).Errorln("failed to GetPushTokens")
		return
	}
	for provider, info := range tokens {
		pusher := u.DefaultPusher
		if p, ok := u.Pushers[provider]; ok {
			pusher = p
		}
		pusher.PushNotif(info, content)
	}
}

func (u *UserSystem) SetPushToken(id appgo.Id, platform appgo.Platform,
	provider, token string) error {
	if provider == "" || token == "" {
		return errors.New("bad push provider or token")
	}
	user := &UserModel{Id: id}
	update := &UserModel{
		Platform:     platform,
		PushProvider: sql.NullString{provider, true},
		PushToken:    sql.NullString{token, true},
	}
	return u.db.Model(user).Updates(update).Error
}

func (u *UserSystem) GetPushTokens(ids []appgo.Id) (map[string]map[appgo.Id]*appgo.PushInfo, error) {
	var users []*UserModel
	if err := u.db.Select("platform, push_provider, push_token").
		Where("id in (?)", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	ret := make(map[string]map[appgo.Id]*appgo.PushInfo)
	for _, user := range users {
		plat, prov, token := user.Platform, user.PushProvider.String, user.PushToken.String
		if plat != 0 && prov != "" && token != "" {
			if _, ok := ret[prov]; !ok {
				ret[prov] = make(map[appgo.Id]*appgo.PushInfo)
			}
			ret[prov][user.Id] = &appgo.PushInfo{
				Platform: plat,
				Provider: prov,
				Token:    token,
			}
		}
	}
	return ret, nil
}
