package redis

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/oxfeeefeee/appgo"
	"time"
)

var (
	pool *redigo.Pool
)

func init() {
	c := &appgo.Conf.Redis
	url := fmt.Sprintf("redis://%s:%s", c.Host, c.Port)
	if c.Password != "" {
		url = fmt.Sprintf("redis://:%s@%s:%s", c.Password, c.Host, c.Port)
	}
	pool = newPool(url, c.MaxIdle, c.IdleTimeout)
}

func Do(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn := pool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
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
