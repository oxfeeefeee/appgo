package userSystem

import (
	"database/sql"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
)

const (
	pushBatchSize = 3
)

func (u *UserSystem) PushTo(users []appgo.Id, content *appgo.PushData) {
	doPush := func(ids []appgo.Id) {
		if tokens, err := u.GetPushTokens(users); err != nil {
			log.WithFields(log.Fields{
				"users": users,
				"error": err,
			}).Errorln("failed to GetPushTokens")
		} else {
			for prov, data := range tokens {
				pusher := u.DefaultPusher
				if p, ok := u.Pushers[prov]; ok {
					pusher = p
				}
				for plat, tokens := range data {
					pusher.PushNotif(plat, tokens, content)
				}
			}
		}
	}
	for i := 0; i < len(users); i += pushBatchSize {
		end := i + pushBatchSize
		if end > len(users) {
			end = len(users)
		}
		doPush(users[i:end])
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

func (u *UserSystem) GetPushTokens(ids []appgo.Id) (map[string]map[appgo.Platform][]string, error) {
	var users []*UserModel
	if err := u.db.Select("platform, push_provider, push_token").
		Where("id in (?)", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	ret := make(map[string]map[appgo.Platform][]string)
	for _, user := range users {
		plat, prov, token := user.Platform, user.PushProvider.String, user.PushToken.String
		if plat != 0 && prov != "" && token != "" {
			if _, ok := ret[prov]; !ok {
				ret[prov] = make(map[appgo.Platform][]string)
			}
			if _, ok := ret[prov][plat]; !ok {
				ret[prov][plat] = make([]string, 0)
			}
			ret[prov][plat] = append(ret[prov][plat], token)
		}
	}
	return ret, nil
}
