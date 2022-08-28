package cache

import (
	"github.com/gomodule/redigo/redis"
)

func Set(key, value string, ttl int64) (err error) {
	_, err = redis.String(Redis.Do("SET", key, value, "EX", ttl))
	return
}

func Get(key string) (value string, err error) {
	return redis.String(Redis.Do("GET", key))
}

func Del(key string) (value string, err error) {
	return redis.String(Redis.Do("DEL", key))
}
