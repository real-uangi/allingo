/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:16
 */

// Package kv

package kv

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/business"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"github.com/redis/go-redis/v9"
	"runtime"
	"time"
)

type RedisKV struct {
	client *redis.Client
	option *redis.Options
	logger *log.StdLogger
}

func newRedisKV(addr, password string) *RedisKV {
	options := &redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           env.GetIntOrDefault("REDIS_DB", 0),
		PoolSize:     env.GetIntOrDefault("REDIS_POOL_SIZE", 10*runtime.GOMAXPROCS(0)),
		MinIdleConns: env.GetIntOrDefault("REDIS_MAX_IDLE", 0),
	}
	return &RedisKV{
		option: options,
		client: redis.NewClient(options),
		logger: log.For[RedisKV](),
	}
}

func (kv *RedisKV) Set(key string, value string, ttl time.Duration) {
	err := filterErr(kv.client.Set(context.Background(), key, value, ttl).Err())
	if err != nil {

	}
}

func (kv *RedisKV) Get(key string) (string, bool) {
	var ok bool
	v, err := kv.client.Get(context.Background(), key).Result()
	if err == nil {
		ok = true
	}
	if err := filterErr(err); err != nil {
		return "", false
	}
	return v, ok
}

func (kv *RedisKV) SetStruct(key string, obj interface{}, ttl time.Duration) error {
	str, err := anyToString(obj)
	if err != nil {
		return err
	}
	kv.Set(key, str, ttl)
	return nil
}

func (kv *RedisKV) GetStruct(key string, p any) error {
	str, ok := kv.Get(key)
	if !ok {
		return nil
	}
	return stringToAny(str, p)
}

func (kv *RedisKV) Del(key string) error {
	return filterErr(kv.client.Del(context.Background(), key).Err())
}

func filterErr(err error) error {
	if err == redis.Nil {
		return nil
	}
	return err
}

func (kv *RedisKV) GetType() Type {
	return Redis
}

func (kv *RedisKV) Ping() error {
	return kv.client.Ping(context.Background()).Err()
}

func (kv *RedisKV) TryLock(key string, parse string, ttl time.Duration) bool {
	b, err := kv.client.SetNX(context.Background(), key, parse, ttl).Result()
	if err = filterErr(err); err != nil {
		kv.logger.Error(err, "error occurs when try lock: "+err.Error())
	}
	return b
}

func (kv *RedisKV) Unlock(key string, parse string) (err error) {
	script := redis.NewScript(`
		if redis.call('get', KEYS[1]) == ARGV[1] 
		then return redis.call('del', KEYS[1])
		else return 0 end;
	`)
	keys := []string{key}
	args := []interface{}{parse}
	result, err := script.Run(context.Background(), kv.client, keys, args).Result()
	if err != nil {
		return err
	}
	if result.(int64) == 0 {
		msg := "unlock failed"
		err = errors.New(msg)
	}
	return
}

func (kv *RedisKV) Incr(key string, i int64, ttl time.Duration, nx bool) (int64, error) {
	result, err := kv.client.IncrBy(context.Background(), key, i).Result()
	if ex := filterErr(err); ex != nil {
		return 0, ex
	}
	if ttl != 0 {
		if nx {
			err = kv.client.ExpireNX(context.Background(), key, ttl).Err()
		} else {
			err = kv.client.Expire(context.Background(), key, ttl).Err()
		}
	}
	return result, filterErr(err)
}

type RedisLock struct {
	kv     *RedisKV
	key    string
	parse  string
	locked bool
}

func (kv *RedisKV) NewLock(key string) Lock {
	return &RedisLock{
		kv:     kv,
		key:    key,
		parse:  uuid.NewString(),
		locked: false,
	}
}

func (lock *RedisLock) TryLock(ttl time.Duration) bool {
	lock.locked = lock.kv.TryLock(lock.key, lock.parse, ttl)
	return lock.locked
}

func (lock *RedisLock) Unlock() error {
	if !lock.locked {
		return nil
	}
	return lock.kv.Unlock(lock.key, lock.parse)
}

func (lock *RedisLock) Lock(ttl, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	for {
		if lock.TryLock(ttl) {
			lock.locked = true
			return nil
		}
		if time.Now().After(deadline) {
			return business.NewError("lock timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
