package redis

import (
	"fmt"
)

type Lists struct {
	namespace string
	client    *Client
}

func NewLists(namespace string, client *Client) *Lists {
	return &Lists{namespace, client}
}

func (l *Lists) keyStr(k interface{}) string {
	return fmt.Sprintf("%s:%v", l.namespace, k)
}

func (l *Lists) AllItems(k interface{}) ([]interface{}, error) {
	if items, err := l.client.Do("LRANGE", l.keyStr(k), 0, -1); err != nil {
		return nil, err
	} else {
		return items.([]interface{}), nil
	}
}

func (l *Lists) LatestItems(k interface{}, num int) ([]interface{}, error) {
	if items, err := l.client.Do("LRANGE", l.keyStr(k), 0, num); err != nil {
		return nil, err
	} else {
		return items.([]interface{}), nil
	}
}

func (l *Lists) Prepend(key interface{}, item interface{}) (int64, error) {
	if newSize, err := l.client.Do("LPUSH", l.keyStr(key), item); err != nil {
		return 0, err
	} else {
		return newSize.(int64), nil
	}
}

func (l *Lists) PopLastItem(key interface{}) (interface{}, error) {
	if item, err := l.client.Do("RPOP", l.keyStr(key)); err != nil {
		return nil, err
	} else {
		return item, nil
	}
}
