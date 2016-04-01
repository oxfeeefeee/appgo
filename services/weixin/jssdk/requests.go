package jssdk

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
	"sync"
	"time"
)

const (
	accessTokenUrl = `https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s`
	ticketUrl      = `https://api.weixin.qq.com/cgi-bin/ticket/getticket?type=jsapi&access_token=%s`
)

var (
	token  *atData
	ticket *ticketData
)

func init() {
	now := time.Now()
	token = &atData{
		expiresAt: now,
		refreshAt: now,
	}
	ticket = &ticketData{
		expiresAt: now,
		refreshAt: now,
	}
}

type atData struct {
	accessToken string
	expiresAt   time.Time
	refreshAt   time.Time
	sync.RWMutex
}

type ticketData struct {
	ticket    string
	expiresAt time.Time
	refreshAt time.Time
	sync.RWMutex
}

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

func getToken() string {
	now := time.Now()
	token.RLock()
	if token.expiresAt.Before(now) {
		token.RUnlock()
		doGetToken()
	} else if token.refreshAt.Before(now) {
		token.RUnlock()
		go doGetToken()
	}
	return token.accessToken
}

func getTicket() string {
	now := time.Now()
	ticket.RLock()
	if ticket.expiresAt.Before(now) {
		ticket.RUnlock()
		doGetTicket()
	} else if ticket.refreshAt.Before(now) {
		ticket.RUnlock()
		go doGetTicket()
	}
	return ticket.ticket
}

func doGetToken() {
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
		return
	}
	var ret accessTokenReturn
	if err := json.Unmarshal(data, &ret); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"body":  string(data),
		}).Error("Failed to unmarshal access token result")
		return
	} else if ret.AccessToken == "" {
		log.WithFields(log.Fields{
			"code": ret.ErrCode,
			"msg":  ret.ErrMsg,
		}).Errorf("weixin token return error")
		return
	}
	token.Lock()
	defer token.Unlock()
	token.accessToken = ret.AccessToken
	token.expiresAt, token.refreshAt = expireConv(ret.ExpiresIn)
}

func doGetTicket() {
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
		return
	}
	var ret ticketReturn
	if err := json.Unmarshal(data, &ret); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"body":  string(data),
		}).Error("Failed to unmarshal weixin ticket result")
		return
	} else if ret.Ticket == "" {
		log.WithFields(log.Fields{
			"code": ret.ErrCode,
			"msg":  ret.ErrMsg,
		}).Errorf("weixin ticket return error")
		return
	}
	ticket.Lock()
	defer ticket.Unlock()
	ticket.ticket = ret.Ticket
	ticket.expiresAt, ticket.refreshAt = expireConv(ret.ExpiresIn)
}

func expireConv(expIn int) (time.Time, time.Time) {
	now := time.Now()
	exp := time.Duration(float32(expIn)*0.9) * time.Second
	ref := time.Duration(float32(expIn)*0.8) * time.Second
	return now.Add(exp), now.Add(ref)
}
