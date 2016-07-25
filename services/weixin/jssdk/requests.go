package jssdk

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo/toolkit/strutil"
	"github.com/parnurzeal/gorequest"
	"time"
)

type accessTokenReturn struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ticketReturn struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	Ticket    string `json:"ticket"`
	ExpiresIn int    `json:"expires_in"`
}

type tokenTicketData struct {
	StoreKey    string
	TokenTicket string
	ExpiresAt   time.Time
	RefreshAt   time.Time
	ExpiresIn   int
}

func (t *tokenTicketData) store() {
	str := strutil.FromObject(t)
	err := kvStore.Set(t.StoreKey, str, t.ExpiresIn)
	if err != nil {
		log.Errorln("kvStore set error: ", err)
	}
}

func readData(key string) *tokenTicketData {
	data, err := kvStore.Get(key)
	if err != nil {
		log.Errorln("kvStore get error: ", err)
		return nil
	}
	if data == "" {
		return nil
	}
	var td tokenTicketData
	strutil.ToObject(&td, data)
	return &td
}

func getToken() string {
	token := readData(tokenStoreKey)
	if token == nil {
		return doGetToken()
	}
	now := time.Now()
	if token.ExpiresAt.Before(now) {
		return doGetToken()
	} else if token.RefreshAt.Before(now) {
		go doGetToken()
	}
	return token.TokenTicket
}

func getTicket() string {
	ticket := readData(ticketStoreKey)
	if ticket == nil {
		return doGetTicket()
	}
	now := time.Now()
	if ticket.ExpiresAt.Before(now) {
		return doGetTicket()
	} else if ticket.RefreshAt.Before(now) {
		go doGetTicket()
	}
	return ticket.TokenTicket
}

func doGetToken() string {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("doGetToken paniced: ", r)
		}
	}()
	url := fmt.Sprintf(accessTokenUrl, appId, appSecret)
	_, data, errs := gorequest.New().Get(url).EndBytes()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    url,
		}).Error("Request weixin token error")
		return ""
	}
	var ret accessTokenReturn
	if err := json.Unmarshal(data, &ret); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"body":  string(data),
		}).Error("Failed to unmarshal access token result")
		return ""
	} else if ret.AccessToken == "" {
		log.WithFields(log.Fields{
			"code": ret.ErrCode,
			"msg":  ret.ErrMsg,
		}).Errorf("weixin token return error")
		return ""
	}
	var token tokenTicketData
	token.StoreKey = tokenStoreKey
	token.TokenTicket = ret.AccessToken
	token.ExpiresIn = ret.ExpiresIn
	token.ExpiresAt, token.RefreshAt = expireConv(ret.ExpiresIn)
	token.store()
	return token.TokenTicket
}

func doGetTicket() string {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("doGetTicket paniced: ", r)
		}
	}()
	token := getToken()
	url := fmt.Sprintf(ticketUrl, token)
	_, data, errs := gorequest.New().Get(url).EndBytes()
	if errs != nil {
		log.WithFields(log.Fields{
			"errors": errs,
			"url":    url,
		}).Error("Request weixin ticket error")
		return ""
	}
	var ret ticketReturn
	if err := json.Unmarshal(data, &ret); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"body":  string(data),
		}).Error("Failed to unmarshal weixin ticket result")
		return ""
	} else if ret.Ticket == "" {
		log.WithFields(log.Fields{
			"code": ret.ErrCode,
			"msg":  ret.ErrMsg,
		}).Errorf("weixin ticket return error")
		return ""
	}
	var ticket tokenTicketData
	ticket.StoreKey = ticketStoreKey
	ticket.TokenTicket = ret.Ticket
	ticket.ExpiresIn = ret.ExpiresIn
	ticket.ExpiresAt, ticket.RefreshAt = expireConv(ret.ExpiresIn)
	ticket.store()
	return ticket.TokenTicket
}

func expireConv(expIn int) (time.Time, time.Time) {
	now := time.Now()
	exp := time.Duration(float32(expIn)*0.9) * time.Second
	ref := time.Duration(float32(expIn)*0.8) * time.Second
	return now.Add(exp), now.Add(ref)
}
