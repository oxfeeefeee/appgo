package appgo

import (
//"strconv"
)

const AnonymousId = 6666

const InternalTestId = 7777

const InternalTestToken = "sjadfjlksadfjkljfwoeifshgsdhgsldfjf"

const CustomTokenHeaderName = "X-Appgo-Token"

const (
	RoleAppUser  Role = 100
	RoleWebUser       = 110
	RoleWebAdmin      = 210
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

type DummyInput struct{}

type KvStore interface {
	Set(k, v string, timeout int) error
	Get(k string) (string, error)
}

type MobileMsgSender interface {
	SendMobileCode(mobile string, template SmsTemplate, code string) error
}
