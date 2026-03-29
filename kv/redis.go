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

var (
	unlockScript = redis.NewScript(`
		if redis.call('get', KEYS[1]) == ARGV[1]
		then return redis.call('del', KEYS[1])
		else return 0 end;
	`)
	compareAndSetScript = redis.NewScript(`
		if redis.call('get', KEYS[1]) == ARGV[1] then
			redis.call('set', KEYS[1], ARGV[2])
			if tonumber(ARGV[3]) > 0 then
				redis.call('pexpire', KEYS[1], ARGV[3])
			end
			return 1
		end
		return 0
	`)
	compareAndDeleteScript = redis.NewScript(`
		if redis.call('get', KEYS[1]) == ARGV[1]
		then return redis.call('del', KEYS[1])
		else return 0 end;
	`)
	getAndSetScript = redis.NewScript(`
		local current = redis.call('get', KEYS[1])
		redis.call('set', KEYS[1], ARGV[1])
		if tonumber(ARGV[2]) > 0 then
			redis.call('pexpire', KEYS[1], ARGV[2])
		end
		if current then
			return {1, current}
		end
		return {0}
	`)
	refreshLockScript = redis.NewScript(`
		if tonumber(ARGV[2]) <= 0 then
			return 0
		end
		if redis.call('get', KEYS[1]) == ARGV[1]
		then return redis.call('pexpire', KEYS[1], ARGV[2])
		else return 0 end;
	`)
)

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

func (kv *RedisKV) SetIfAbsent(key string, value string, ttl time.Duration) (bool, error) {
	return kv.client.SetNX(context.Background(), key, value, ttl).Result()
}

func (kv *RedisKV) CompareAndSet(key string, expected, value string, ttl time.Duration) (bool, error) {
	result, err := compareAndSetScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		expected,
		value,
		ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (kv *RedisKV) CompareAndDelete(key string, expected string) (bool, error) {
	result, err := compareAndDeleteScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		expected,
	).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (kv *RedisKV) GetAndDelete(key string) (string, bool, error) {
	value, err := kv.client.GetDel(context.Background(), key).Result()
	if err == nil {
		return value, true, nil
	}
	if err := filterErr(err); err != nil {
		return "", false, err
	}
	return "", false, nil
}

func (kv *RedisKV) GetAndSet(key string, value string, ttl time.Duration) (string, bool, error) {
	result, err := getAndSetScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		value,
		ttl.Milliseconds(),
	).Result()
	if err != nil {
		return "", false, err
	}
	items, ok := result.([]interface{})
	if !ok || len(items) == 0 {
		return "", false, errors.New("unexpected get-and-set result")
	}
	flag, ok := items[0].(int64)
	if !ok {
		return "", false, errors.New("unexpected get-and-set flag")
	}
	if flag == 0 {
		return "", false, nil
	}
	if len(items) < 2 {
		return "", false, errors.New("unexpected get-and-set payload")
	}
	switch old := items[1].(type) {
	case string:
		return old, true, nil
	case []byte:
		return string(old), true, nil
	default:
		return "", false, errors.New("unexpected get-and-set value")
	}
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
	result, err := unlockScript.Run(context.Background(), kv.client, []string{key}, parse).Result()
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
	err := lock.kv.Unlock(lock.key, lock.parse)
	if err == nil {
		lock.locked = false
	}
	return err
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

func (lock *RedisLock) Refresh(ttl time.Duration) (bool, error) {
	if !lock.locked {
		return false, nil
	}
	result, err := refreshLockScript.Run(
		context.Background(),
		lock.kv.client,
		[]string{lock.key},
		lock.parse,
		ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return false, err
	}
	ok := result == 1
	if !ok {
		lock.locked = false
	}
	return ok, nil
}
