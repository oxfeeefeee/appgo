package userSystem

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
)

type defaultKvStore map[string]string

func (kv defaultKvStore) Set(k, v string, timeout int) error {
	kv[k] = v
	return nil
}

func (kv defaultKvStore) Get(k string) (string, error) {
	v := kv[k]
	return v, nil
}

type defaultSender struct{}

func (s defaultSender) SendMobileCode(mobile string, template appgo.SmsTemplate, code string) error {
	log.Infoln("SendMobileCode: ", mobile, ", ", template, ", ", code)
	return nil
}

type defaultPusher struct{}

func (s defaultPusher) Name() string {
	return "default"
}

func (s defaultPusher) PushNotif(
	platform appgo.Platform, tokens []string, content *appgo.PushData) {
	log.Infoln("PushNotif: ", platform, ", ", tokens, ", ", content)
}
