package redis

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
)

type collection struct {
	namespace string
	expire    int
}

func newCollection(namespace string, expire int) *collection {
	return &collection{"cl:" + namespace, expire}
}

func (c *collection) has(key interface{}) (bool, error) {
	if has, err := redigo.Bool(Do("EXISTS", c.ckey(key))); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (c *collection) set(key interface{}, val interface{}) error {
	if c.expire > 0 {
		if _, err := Do("SETEX", c.ckey(key), c.expire, val); err != nil {
			return err
		}
	} else {
		if _, err := Do("SET", c.ckey(key), val); err != nil {
			return err
		}
	}
	return nil
}

// overwrite default expire
func (c *collection) setEx(key interface{}, expire int, val interface{}) error {
	if _, err := Do("SETEX", c.ckey(key), expire, val); err != nil {
		return err
	}
	return nil
}

func (c *collection) get(key interface{}) (interface{}, error) {
	return Do("GET", c.ckey(key))
}

func (c *collection) ckey(k interface{}) string {
	return fmt.Sprintf("%s:%v", c.namespace, k)
}
