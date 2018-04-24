package redis

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/oxfeeefeee/appgo"
	"time"
)

var (
	BaseClient    *Client
	MetricsClient *Client
	ErrNotFound   = appgo.NewApiErr(appgo.ECodeNotFound, "redis: value not found")
)

type Client struct {
	Pool *redigo.Pool
}

func NewClient(conf appgo.RedisConf) *Client {
	url := fmt.Sprintf("redis://%s:%s", conf.Host, conf.Port)
	if conf.Password != "" {
		url = fmt.Sprintf("redis://:%s@%s:%s", conf.Password, conf.Host, conf.Port)
	}
	pool := newPool(url, conf.MaxIdle, conf.IdleTimeout)
	return &Client{Pool: pool}
}

type Trans struct {
	conn redigo.Conn
}

func (t *Trans) Send(cmd string, args ...interface{}) {
	t.conn.Send(cmd, args...)
}

func (t *Trans) Exec() (reply interface{}, err error) {
	defer t.conn.Close()
	return t.conn.Do("EXEC")
}

func (self *Client) Do(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn := self.Pool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func (self *Client) GetConn() redigo.Conn {
	return self.Pool.Get()
}

func (self *Client) BeginTrans() *Trans {
	conn := self.Pool.Get()
	conn.Send("MULTI")
	return &Trans{conn}
}

func newPool(url string, maxIdle, idleTimeout int) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.DialURL(url)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
