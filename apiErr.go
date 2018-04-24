package appgo

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

var (
	NotFoundErr                    *ApiError
	UnauthorizedErr                *ApiError
	ForbiddenErr                   *ApiError
	InternalErr                    *ApiError
	InvalidUsernameErr             *ApiError
	InvalidNicknameErr             *ApiError
	InvalidPasswordErr             *ApiError
	PasswordNotSetErr              *ApiError
	MobileUserNotFoundErr          *ApiError
	MobileUserBadCodeErr           *ApiError
	MobileUserBadTokenErr          *ApiError
	MobileUserAlreadyExistsErr     *ApiError
	WeixinUserAlreadyExistsErr     *ApiError
	UserWithoutMobileErr           *ApiError
	WeixinUserWithAnotherMobileErr *ApiError
)

const (
	ECodeOK                          ErrCode = 20000
	ECodeRedirect                            = 30200
	ECodeBadRequest                          = 40000
	ECodeUnauthorized                        = 40100
	ECodeForbidden                           = 40300
	ECodeNotFound                            = 40400
	ECodeInternal                            = 50000
	ECode3rdPartyAuthFailed                  = 50300
	ECodeInvalidUsername                     = 60001
	ECodeInvalidNickname                     = 60002
	ECodeInvalidPassword                     = 60003
	ECodePasswordNotSet                      = 60004
	ECodeMobileUserNotFound                  = 60101
	ECodeMobileUserBadCode                   = 60102
	ECodeMobileUserBadToken                  = 60103
	ECodeMobileUserAlreadyExists             = 60104
	ECodeWeixinUserAlreadyExists             = 60107
	ECodeWeixinUserWithAnotherMobile         = 60108
	ECodeUserWithoutMobile                   = 60109
)

type ErrCode int

func init() {
	NotFoundErr = NewApiErr(ECodeNotFound, "NotFound error")
	UnauthorizedErr = NewApiErr(ECodeUnauthorized, "Unauthorized error")
	ForbiddenErr = NewApiErr(ECodeForbidden, "Forbidden error")
	InternalErr = NewApiErr(ECodeInternal, "Internal error")
	InvalidUsernameErr = NewApiErr(ECodeInvalidUsername, "Invalid username")
	InvalidNicknameErr = NewApiErr(ECodeInvalidNickname, "Invalid nickname")
	InvalidPasswordErr = NewApiErr(ECodeInvalidPassword, "Invalid password")
	PasswordNotSetErr = NewApiErr(ECodePasswordNotSet, "Password not set yet")
	MobileUserNotFoundErr = NewApiErr(ECodeMobileUserNotFound, "Mobile user not found")
	MobileUserBadCodeErr = NewApiErr(ECodeMobileUserBadCode, "Mobile user bad code")
	MobileUserBadTokenErr = NewApiErr(ECodeMobileUserBadToken, "Mobile user bad token")
	MobileUserAlreadyExistsErr = NewApiErr(ECodeMobileUserAlreadyExists, "Mobile user already exists")
	WeixinUserAlreadyExistsErr = NewApiErr(ECodeWeixinUserAlreadyExists, "Weixin user already exists")
	UserWithoutMobileErr = NewApiErr(ECodeUserWithoutMobile, "User without mobile")
	WeixinUserWithAnotherMobileErr = NewApiErr(ECodeWeixinUserWithAnotherMobile, "Weixin user with another mobile")
}

type ApiError struct {
	Code ErrCode `json:"errcode"`
	Msg  string  `json:"errmsg"`
}

func (e *ApiError) Error() string {
	return e.Msg
}

func (e *ApiError) HttpCode() int {
	return int(e.Code) / 100
}

func (e *ApiError) HttpError(w http.ResponseWriter) {
	code := 200 //int(e.Code) / 100
	http.Error(w, "", code)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(e)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"ApiError": e,
		}).Error("Failed to encode ApiError")
	}
}

func NewApiErr(code ErrCode, msg string) *ApiError {
	return &ApiError{code, msg}
}

func NewApiErrWithCode(code ErrCode) *ApiError {
	return &ApiError{code, "No extra info"}
}

func NewApiErrWithMsg(msg string) *ApiError {
	return &ApiError{ECodeInternal, msg}
}

func ApiErrFromGoErr(err error) *ApiError {
	if e, ok := err.(*ApiError); ok {
		return e
	} else {
		return NewApiErr(ECodeInternal, err.Error())
	}
}
