package redis

import (
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.zerodha.tech/commons/lil/store"
)

var (
	mockRedis *miniredis.Miniredis
)

func init() {
	var err error
	mockRedis, err = miniredis.Run()
	if err != nil {
		panic(err)
	}
}

func getRedisPool() *redis.Pool {
	return &redis.Pool{
		Wait: true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				mockRedis.Addr(),
			)

			return c, err
		},
	}
}

func TestSet(t *testing.T) {
	assert := assert.New(t)
	str := New(getRedisPool())
	var (
		err  error
		id   = "someid"
		url  = "http://example.com"
		meta = &store.Meta{}
	)
	err = str.Set(id, url, meta)
	assert.NoError(err)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	str := New(getRedisPool())
	var (
		err  error
		id   = "someid"
		url  = "http://example.com"
		meta = &store.Meta{}
	)
	err = str.Set(id, url, meta)
	assert.NoError(err)
	rURL, _, err := str.Get(id)
	assert.NoError(err)
	assert.Equal(url, rURL)
}

func TestGetErrorNotFound(t *testing.T) {
	assert := assert.New(t)
	str := New(getRedisPool())
	_, _, err := str.Get("_some_random_id")
	assert.Error(err, store.ErrNotFound.Error())
}

func TestID(t *testing.T) {
	assert := assert.New(t)
	str := New(getRedisPool())
	var (
		err  error
		id   = "someid"
		url  = "http://example.com"
		meta = &store.Meta{}
	)
	err = str.Set(id, url, meta)
	assert.NoError(err)
	rID, err := str.GetID(url)
	assert.NoError(err)
	assert.Equal(id, rID)
}

func TestIDErrorNotFound(t *testing.T) {
	assert := assert.New(t)
	str := New(getRedisPool())
	_, err := str.GetID("http://randomurl.com")
	assert.Error(err, store.ErrNotFound.Error())
}

func TestDeleteErrorNotFound(t *testing.T) {
	assert := assert.New(t)
	str := New(getRedisPool())
	err := str.Del("_some_random_id")
	assert.Error(err, store.ErrNotFound.Error())
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	str := New(getRedisPool())
	var (
		err  error
		id   = "someid"
		url  = "http://example.com"
		meta = &store.Meta{}
	)
	err = str.Set(id, url, meta)
	assert.NoError(err)
	err = str.Del(id)
	assert.NoError(err)
	_, _, err = str.Get(id)
	assert.Error(err, store.ErrNotFound.Error())
}
