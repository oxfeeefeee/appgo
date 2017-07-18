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

var (
	ErrNotFound = appgo.NewApiErr(appgo.ECodeNotFound, "redis: value not found")
)

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

func init() {
	c := &appgo.Conf.Redis[0]
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

func GetConn() redigo.Conn {
	return pool.Get()
}

func BeginTrans() *Trans {
	conn := pool.Get()
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
