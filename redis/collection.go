package redis

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
)

type DataStore interface {
	Read(key interface{}) (interface{}, error)
	Write(key, val interface{}) error
}

type Serializer interface {
	ToBytes(val interface{}) ([]byte, error)
	FromBytes(data []byte, val interface{}) error
}

type Collection struct {
	namespace  string
	expire     int
	dataStore  DataStore
	serializer Serializer
}

func NewCollection(namespace string, expire int, ds DataStore, sl Serializer) *Collection {
	return &Collection{"cl:" + namespace,
		expire,
		ds,
		sl,
	}
}

func (c *Collection) Has(key interface{}) (bool, error) {
	if has, err := redigo.Bool(Do("EXISTS", c.ckey(key))); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (c *Collection) Set(key interface{}, val interface{}) error {
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
func (c *Collection) SetEx(key interface{}, expire int, val interface{}) error {
	if _, err := Do("SETEX", c.ckey(key), expire, val); err != nil {
		return err
	}
	return nil
}

func (c *Collection) SetFromStore(key interface{}) error {
	if val, err := c.dataStore.Read(key); err != nil {
		return err
	} else if bs, err := c.serializer.ToBytes(val); err != nil {
		return err
	} else {
		return c.Set(key, bs)
	}
}

func (c *Collection) GetString(key interface{}) (string, error) {
	if val, err := Do("GET", c.ckey(key)); err != nil {
		return "", err
	} else if val == nil && c.dataStore != nil {
		if v, err := c.dataStore.Read(key); err != nil {
			return "", err
		} else {
			str := v.(string)
			if err := c.Set(key, str); err != nil {
				return "", err
			}
			return str, nil
		}
	} else {
		return redigo.String(val, nil)
	}
}

func (c *Collection) GetObj(key, val interface{}) error {
	var data []byte
	if val, err := Do("GET", c.ckey(key)); err != nil {
		return err
	} else if val == nil && c.dataStore != nil {
		if v, err := c.dataStore.Read(key); err != nil {
			return err
		} else if bs, err := c.serializer.ToBytes(v); err != nil {
			return err
		} else {
			if err := c.Set(key, bs); err != nil {
				return err
			}
			data = bs
		}
	} else if bs, err := redigo.Bytes(val, nil); err != nil {
		return err
	} else {
		data = bs
	}
	return c.serializer.FromBytes(data, val)
}

func (c *Collection) ckey(k interface{}) string {
	return fmt.Sprintf("%s:%v", c.namespace, k)
}
