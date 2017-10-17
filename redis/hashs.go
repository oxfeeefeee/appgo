package redis

import (
	redigo "github.com/garyburd/redigo/redis"
)

type Hashs struct {
	namespace string
}

func NewHashs(namespace string) *Hashs {
	return &Hashs{namespace}
}

func (h *Hashs) SetValue(field interface{}, value interface{}) (int64, error) {
	if result, err := Do("HSET", h.namespace, field, value); err != nil {
		return 0, err
	} else {
		return result.(int64), nil
	}
}

func (h *Hashs) GetValue(field interface{}) (string, error) {
	if result, err := Do("HGET", h.namespace, field); err != nil {
		return "", err
	} else {
		return redigo.String(result, nil)
	}
}
