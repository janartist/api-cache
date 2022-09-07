package store

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

type RedisConf struct {
	Addr string `yaml:"addr"`
	Auth string `yaml:"auth"`
	DB   int    `yaml:"db"`
}

type RedisStore struct {
	*redis.Client
}

func NewRedisStore(opt *redis.Options) *RedisStore {
	client := redis.NewClient(opt)
	result, err := client.Ping().Result()
	if err != nil || result != "PONG" {
		panic(err)
	}
	return &RedisStore{client}
}
func NewRedisStoreDefault(conf *RedisConf) *RedisStore {
	auth, err := base64.StdEncoding.DecodeString(conf.Auth)
	if err != nil {
		panic(err)
	}
	opt, err := redis.ParseURL(fmt.Sprintf("redis://:%s@%s/%d", auth, conf.Addr, conf.DB))
	if err != nil {
		panic(err)
	}
	opt.ReadTimeout = time.Second * 2
	opt.WriteTimeout = time.Second * 2
	opt.PoolSize = 10

	return NewRedisStore(opt)
}

func (c *RedisStore) Set(key string, k string, val *ResponseCache, expires time.Duration) error {
	v, err := json.Marshal(val)
	if err != nil {
		return err
	}
	_, err = c.Client.HSet(key, k, v).Result()
	_, err = c.Client.Expire(key, expires).Result()
	return err
}

func (c *RedisStore) Get(key string, k string, val *ResponseCache) error {
	r, err := c.Client.HGet(key, k).Result()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(r), val)
	if err != nil {
		return err
	}
	e, err := c.Client.TTL(key).Result()
	if err != nil {
		return err
	}
	val.Ttl = e
	return nil
}

func (c *RedisStore) Remove(key string) error {
	_, err := c.Client.Del(key).Result()
	return err
}
