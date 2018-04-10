package redis

import (
	redigo "github.com/garyburd/redigo/redis"
)

type Hashs struct {
	namespace string
	client    *Client
}

func NewHashs(namespace string, client *Client) *Hashs {
	return &Hashs{namespace, client}
}

func (h *Hashs) SetValue(field interface{}, value interface{}) (int64, error) {
	if result, err := h.client.Do("HSET", h.namespace, field, value); err != nil {
		return 0, err
	} else {
		return result.(int64), nil
	}
}

func (h *Hashs) GetValue(field interface{}) (string, error) {
	if result, err := h.client.Do("HGET", h.namespace, field); err != nil {
		if err == ErrNotFound {
			return "", nil
		}
		return "", err
	} else {
		if result == nil {
			return "", nil
		}
		return redigo.String(result, nil)
	}
}
