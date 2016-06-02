package appgo

import (
//"strconv"
)

const AnonymousId = 6666

const InternalTestId = 7777

const InternalTestToken = "sjadfjlksadfjkljfwoeifshgsdhgsldfjf"

const CustomTokenHeaderName = "X-Appgo-Token"

const CustomVersionHeaderName = "X-Appgo-Api-Version"

const (
	RoleAppUser  Role = 100
	RoleWebUser       = 101
	RoleWebAdmin      = 200
)

type Role int

const (
	SexDefault Sex = iota
	SexMale
	SexFemale
)

type Sex int8

const (
	_ SmsTemplate = iota
	SmsTemplateRegister
	SmsTemplatePwReset
)

type SmsTemplate int

const (
	_ Platform = iota
	PlatformIos
	PlatformAndroid
)

type Platform int8

type DummyInput struct{}

type KvStore interface {
	Set(k, v string, timeout int) error
	Get(k string) (string, error)
}

type MobileMsgSender interface {
	SendMobileCode(mobile string, template SmsTemplate, code string) error
}

type PushInfo struct {
	Platform Platform
	Provider string
	Token    string
}

type PushData struct {
	Title       string
	Message     string
	Sound       string
	Badge       int
	PlayVibrate bool
	PlayLights  bool
	PlaySound   bool
	Custom      map[string]interface{}
}

type Pusher interface {
	Name() string
	PushNotif(platform Platform, tokens []string, content *PushData)
}

func init() {
	initConfig()
	initLog()
}
