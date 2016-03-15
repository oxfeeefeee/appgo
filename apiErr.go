package appgo

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

var (
	NotFoundErr     error
	UnauthorizedErr error
	ForbiddenErr    error
	InternalErr     error
)

func init() {
	NotFoundErr = NewApiErrWithCode(ECodeNotFound)
	UnauthorizedErr = NewApiErrWithCode(ECodeUnauthorized)
	ForbiddenErr = NewApiErrWithCode(ECodeForbidden)
	InternalErr = NewApiErrWithCode(ECodeInternal)
}

type ApiError struct {
	Code ErrCode `json:"errcode"`
	Msg  string  `json:"msg"`
}

func (e *ApiError) Error() string {
	return e.Msg
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
	return &ApiError{code, ""}
}
