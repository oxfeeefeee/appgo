package appgo

import (
	"strconv"
)

const (
	RoleAppUser  Role = 100
	RoleWebUser       = 110
	RoleWebAdmin      = 210
)

type Role int

const (
	ECodeOK                 ErrCode = 20000
	ECodeBadRequest                 = 40000
	ECodeUnauthorized               = 40100
	ECodeForbidden                  = 40300
	ECodeNotFound                   = 40400
	ECodeInternal                   = 50000
	ECode3rdPartyAuthFailed         = 50300
)

type ErrCode int

type DummyInput struct{}

type Id int64

func (id Id) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func IdFromStr(str string) Id {
	i, _ := strconv.Atoi(str)
	return Id(i)
}
