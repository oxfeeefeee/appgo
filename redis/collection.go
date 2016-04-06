package redis

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
)

type Collection struct {
	namespace string
	reader    ObjReader
}

func NewCollection(namespace string, reader ObjReader) *Collection {
	return &Collection{"cl:" + namespace, reader}
}

func (c *Collection) Has(key interface{}) (bool, error) {
	if has, err := redigo.Bool(Do("EXISTS", c.ckey(key))); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (c *Collection) Set(key interface{}, val string) error {
	if _, err := Do("SET", c.ckey(key), val); err != nil {
		return err
	}
	return nil
}

func (c *Collection) SetEx(key interface{}, expire int, val string) error {
	if _, err := Do("SETEX", c.ckey(key), expire, val); err != nil {
		return err
	}
	return nil
}

func (c *Collection) Get(key interface{}) (string, error) {
	if val, err := Do("GET", c.ckey(key)); err != nil {
		return "", err
	} else if val == nil && c.reader != nil {
		if v, err := c.reader(key); err != nil {
			return "", err
		} else {
			c.Set(key, v)
			return v, nil
		}
	} else {
		return redigo.String(val, nil)
	}
}

func (c *Collection) ckey(k interface{}) string {
	return fmt.Sprintf("%s:%v", c.namespace, k)
}
