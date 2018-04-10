package redis

import (
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

type Objects struct {
	col        collection
	dataStore  DataStore
	serializer Serializer
}

func NewObjects(namespace string, expire int, client *Client, ds DataStore, sl Serializer) *Objects {
	return &Objects{*newCollection(namespace, expire, client), ds, sl}
}

func (o *Objects) Set(key, val interface{}) error {
	if bs, err := o.serializer.ToBytes(val); err != nil {
		return err
	} else {
		return o.col.set(key, bs)
	}
}

func (o *Objects) Get(key, val interface{}) error {
	if v, err := o.col.get(key); err != nil {
		return err
	} else if v == nil {
		return ErrNotFound
	} else if bs, err := redigo.Bytes(val, nil); err != nil {
		return err
	} else {
		return o.serializer.FromBytes(bs, val)
	}
}

func (o *Objects) SetWithStore(key interface{}) error {
	if val, err := o.dataStore.Read(key); err != nil {
		return err
	} else {
		return o.Set(key, val)
	}
}

func (o *Objects) GetWithStore(key, val interface{}) error {
	err := o.Get(key, val)
	if err != ErrNotFound {
		return err
	}
	if v, err := o.dataStore.Read(key); err != nil {
		return err
	} else if bs, err := o.serializer.ToBytes(v); err != nil {
		return err
	} else {
		if err := o.Set(key, bs); err != nil {
			return err
		}
		return o.Get(key, val)
	}
}
