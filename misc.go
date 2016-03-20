package appgo

import (
//"strconv"
)

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

type DummyInput struct{}
