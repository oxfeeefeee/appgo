package redis

import (
	redigo "github.com/garyburd/redigo/redis"
)

type Strings struct {
	col collection
}

func NewStrings(namespace string, expire int, client *Client) *Strings {
	return &Strings{*newCollection(namespace, expire, client)}
}

func (s *Strings) Has(key interface{}) (bool, error) {
	return s.col.has(key)
}

func (s *Strings) Set(key interface{}, val string) error {
	return s.col.set(key, val)
}

// overwrite default expire
func (s *Strings) SetEx(key interface{}, expire int, val string) error {
	return s.col.setEx(key, expire, val)
}

func (s *Strings) Get(key interface{}) (string, error) {
	val, err := s.col.get(key)
	if err != nil {
		return "", err
	} else if val == nil {
		return "", ErrNotFound
	}
	return redigo.String(val, nil)
}
