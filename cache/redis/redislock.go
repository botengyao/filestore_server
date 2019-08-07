package redis

import (
	"github.com/garyburd/redigo/redis"
)

const (
	redisLockTimeout = 2 // 1 seconds
)

func Lock(con redis.Conn, key string) (isLock bool, err error) {
	//这里需要redis.String包一下，才能返回redis.ErrNil
	_, err = redis.String(con.Do("set", key, 1, "ex", redisLockTimeout, "nx"))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		return
	}
	isLock = true
	return
}

func Unlock(con redis.Conn, key string) (err error) {
	_, err = con.Do("del", key)
	if err != nil {
		return
	}
	return
}
