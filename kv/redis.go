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
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/business"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"github.com/redis/go-redis/v9"
)

const redisWrongTypeMessage = "WRONGTYPE Operation against a key holding the wrong kind of value"

type RedisKV struct {
	client *redis.Client
	option *redis.Options
	logger *log.StdLogger
}

var (
	stringSetScript = redis.NewScript(`
		local kind = redis.call('type', KEYS[1]).ok
		if kind ~= 'none' and kind ~= 'string' then
			return redis.error_reply(ARGV[3])
		end
		redis.call('set', KEYS[1], ARGV[1])
		if tonumber(ARGV[2]) > 0 then
			redis.call('pexpire', KEYS[1], ARGV[2])
		else
			redis.call('persist', KEYS[1])
		end
		return 1
	`)
	stringSetIfAbsentScript = redis.NewScript(`
		local kind = redis.call('type', KEYS[1]).ok
		if kind ~= 'none' and kind ~= 'string' then
			return redis.error_reply(ARGV[3])
		end
		if kind == 'string' then
			return 0
		end
		redis.call('set', KEYS[1], ARGV[1])
		if tonumber(ARGV[2]) > 0 then
			redis.call('pexpire', KEYS[1], ARGV[2])
		else
			redis.call('persist', KEYS[1])
		end
		return 1
	`)
	stringSetIfPresentScript = redis.NewScript(`
		local kind = redis.call('type', KEYS[1]).ok
		if kind == 'none' then
			return 0
		end
		if kind ~= 'string' then
			return redis.error_reply(ARGV[3])
		end
		local currentTTL = redis.call('pttl', KEYS[1])
		redis.call('set', KEYS[1], ARGV[1])
		if tonumber(ARGV[2]) > 0 then
			redis.call('pexpire', KEYS[1], ARGV[2])
		elseif currentTTL > 0 then
			redis.call('pexpire', KEYS[1], currentTTL)
		else
			redis.call('persist', KEYS[1])
		end
		return 1
	`)
	compareAndSetScript = redis.NewScript(`
		local kind = redis.call('type', KEYS[1]).ok
		if kind == 'none' then
			return 0
		end
		if kind ~= 'string' then
			return redis.error_reply(ARGV[4])
		end
		local current = redis.call('get', KEYS[1])
		if current ~= ARGV[1] then
			return 0
		end
		local currentTTL = redis.call('pttl', KEYS[1])
		redis.call('set', KEYS[1], ARGV[2])
		if tonumber(ARGV[3]) > 0 then
			redis.call('pexpire', KEYS[1], ARGV[3])
		elseif currentTTL > 0 then
			redis.call('pexpire', KEYS[1], currentTTL)
		else
			redis.call('persist', KEYS[1])
		end
		return 1
	`)
	compareAndDeleteScript = redis.NewScript(`
		local kind = redis.call('type', KEYS[1]).ok
		if kind == 'none' then
			return 0
		end
		if kind ~= 'string' then
			return redis.error_reply(ARGV[2])
		end
		if redis.call('get', KEYS[1]) == ARGV[1] then
			return redis.call('del', KEYS[1])
		end
		return 0
	`)
	getAndDeleteScript = redis.NewScript(`
		local kind = redis.call('type', KEYS[1]).ok
		if kind == 'none' then
			return {0}
		end
		if kind ~= 'string' then
			return redis.error_reply(ARGV[1])
		end
		local current = redis.call('get', KEYS[1])
		redis.call('del', KEYS[1])
		return {1, current}
	`)
	getAndSetScript = redis.NewScript(`
		local kind = redis.call('type', KEYS[1]).ok
		if kind ~= 'none' and kind ~= 'string' then
			return redis.error_reply(ARGV[3])
		end
		if kind == 'none' then
			redis.call('set', KEYS[1], ARGV[1])
			if tonumber(ARGV[2]) > 0 then
				redis.call('pexpire', KEYS[1], ARGV[2])
			else
				redis.call('persist', KEYS[1])
			end
			return {0}
		end
		local current = redis.call('get', KEYS[1])
		local currentTTL = redis.call('pttl', KEYS[1])
		redis.call('set', KEYS[1], ARGV[1])
		if tonumber(ARGV[2]) > 0 then
			redis.call('pexpire', KEYS[1], ARGV[2])
		elseif currentTTL > 0 then
			redis.call('pexpire', KEYS[1], currentTTL)
		else
			redis.call('persist', KEYS[1])
		end
		return {1, current}
	`)
	unlockScript = redis.NewScript(`
		if redis.call('get', KEYS[1]) == ARGV[1]
		then return redis.call('del', KEYS[1])
		else return 0 end;
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

func normalizeRedisErr(err error) error {
	if err == nil || err == redis.Nil {
		return nil
	}
	if strings.HasPrefix(err.Error(), "WRONGTYPE") {
		return fmt.Errorf("%w: %v", ErrWrongType, err)
	}
	return err
}

func (kv *RedisKV) ensureStringKey(key string) error {
	kind, err := kv.client.Type(context.Background(), key).Result()
	if err != nil {
		return normalizeRedisErr(err)
	}
	if kind == "none" || kind == "string" {
		return nil
	}
	return ErrWrongType
}

func parseScriptBoolResult(result any) (bool, error) {
	switch value := result.(type) {
	case int64:
		return value == 1, nil
	default:
		return false, errors.New("unexpected script result")
	}
}

func parseStringScriptResult(result any) (string, bool, error) {
	items, ok := result.([]interface{})
	if !ok || len(items) == 0 {
		return "", false, errors.New("unexpected string script result")
	}
	flag, ok := items[0].(int64)
	if !ok {
		return "", false, errors.New("unexpected string script flag")
	}
	if flag == 0 {
		return "", false, nil
	}
	if len(items) < 2 {
		return "", false, errors.New("unexpected string script payload")
	}
	switch value := items[1].(type) {
	case string:
		return value, true, nil
	case []byte:
		return string(value), true, nil
	default:
		return "", false, errors.New("unexpected string script value")
	}
}

func (kv *RedisKV) refreshTTL(key string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return normalizeRedisErr(kv.client.Expire(context.Background(), key, ttl).Err())
}

func (kv *RedisKV) Set(key string, value string, ttl time.Duration) error {
	_, err := stringSetScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		value,
		ttl.Milliseconds(),
		redisWrongTypeMessage,
	).Result()
	return normalizeRedisErr(err)
}

func (kv *RedisKV) Get(key string) (string, bool, error) {
	value, err := kv.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, normalizeRedisErr(err)
	}
	return value, true, nil
}

func (kv *RedisKV) Exists(key string) (bool, error) {
	count, err := kv.client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return count > 0, nil
}

func (kv *RedisKV) Del(key string) (bool, error) {
	count, err := kv.client.Del(context.Background(), key).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return count > 0, nil
}

func (kv *RedisKV) Expire(key string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, nil
	}
	ok, err := kv.client.Expire(context.Background(), key, ttl).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return ok, nil
}

func (kv *RedisKV) Persist(key string) (bool, error) {
	ok, err := kv.client.Persist(context.Background(), key).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return ok, nil
}

func (kv *RedisKV) TTL(key string) (time.Duration, bool, error) {
	ttl, err := kv.client.TTL(context.Background(), key).Result()
	if err != nil {
		return 0, false, normalizeRedisErr(err)
	}
	switch {
	case ttl == time.Duration(-2):
		return 0, false, nil
	case ttl == time.Duration(-1):
		return 0, true, nil
	default:
		return ttl, true, nil
	}
}

func (kv *RedisKV) SetIfAbsent(key string, value string, ttl time.Duration) (bool, error) {
	result, err := stringSetIfAbsentScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		value,
		ttl.Milliseconds(),
		redisWrongTypeMessage,
	).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return parseScriptBoolResult(result)
}

func (kv *RedisKV) SetIfPresent(key string, value string, ttl time.Duration) (bool, error) {
	result, err := stringSetIfPresentScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		value,
		ttl.Milliseconds(),
		redisWrongTypeMessage,
	).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return parseScriptBoolResult(result)
}

func (kv *RedisKV) CompareAndSet(key string, expected, value string, ttl time.Duration) (bool, error) {
	result, err := compareAndSetScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		expected,
		value,
		ttl.Milliseconds(),
		redisWrongTypeMessage,
	).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return parseScriptBoolResult(result)
}

func (kv *RedisKV) CompareAndDelete(key string, expected string) (bool, error) {
	result, err := compareAndDeleteScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		expected,
		redisWrongTypeMessage,
	).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return parseScriptBoolResult(result)
}

func (kv *RedisKV) GetAndDelete(key string) (string, bool, error) {
	result, err := getAndDeleteScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		redisWrongTypeMessage,
	).Result()
	if err != nil {
		return "", false, normalizeRedisErr(err)
	}
	return parseStringScriptResult(result)
}

func (kv *RedisKV) GetAndSet(key string, value string, ttl time.Duration) (string, bool, error) {
	result, err := getAndSetScript.Run(
		context.Background(),
		kv.client,
		[]string{key},
		value,
		ttl.Milliseconds(),
		redisWrongTypeMessage,
	).Result()
	if err != nil {
		return "", false, normalizeRedisErr(err)
	}
	return parseStringScriptResult(result)
}

func (kv *RedisKV) MGet(keys ...string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		value, ok, err := kv.Get(key)
		if err != nil {
			return nil, err
		}
		if ok {
			result[key] = value
		}
	}
	return result, nil
}

func (kv *RedisKV) MSet(values map[string]string, ttl time.Duration) error {
	for key := range values {
		if err := kv.ensureStringKey(key); err != nil {
			return err
		}
	}
	for key, value := range values {
		if err := kv.Set(key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

func (kv *RedisKV) MDel(keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	count, err := kv.client.Del(context.Background(), keys...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return count, nil
}

func (kv *RedisKV) SetStruct(key string, obj any, ttl time.Duration) error {
	str, err := anyToString(obj)
	if err != nil {
		return err
	}
	return kv.Set(key, str, ttl)
}

func (kv *RedisKV) GetStruct(key string, p any) (bool, error) {
	str, ok, err := kv.Get(key)
	if err != nil || !ok {
		return ok, err
	}
	return true, stringToAny(str, p)
}

func (kv *RedisKV) HSet(key string, field string, value string, ttl time.Duration) (bool, error) {
	count, err := kv.client.HSet(context.Background(), key, field, value).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	if err := kv.refreshTTL(key, ttl); err != nil {
		return false, err
	}
	return count == 1, nil
}

func (kv *RedisKV) HSetIfAbsent(key string, field string, value string, ttl time.Duration) (bool, error) {
	ok, err := kv.client.HSetNX(context.Background(), key, field, value).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	if err := kv.refreshTTL(key, ttl); err != nil {
		return false, err
	}
	return ok, nil
}

func (kv *RedisKV) HGet(key string, field string) (string, bool, error) {
	value, err := kv.client.HGet(context.Background(), key, field).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, normalizeRedisErr(err)
	}
	return value, true, nil
}

func (kv *RedisKV) HDel(key string, fields ...string) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}
	count, err := kv.client.HDel(context.Background(), key, fields...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return count, nil
}

func (kv *RedisKV) HExists(key string, field string) (bool, error) {
	ok, err := kv.client.HExists(context.Background(), key, field).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return ok, nil
}

func (kv *RedisKV) HGetAll(key string) (map[string]string, error) {
	result, err := kv.client.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return result, nil
}

func (kv *RedisKV) HLen(key string) (int64, error) {
	count, err := kv.client.HLen(context.Background(), key).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return count, nil
}

func (kv *RedisKV) HIncr(key string, field string, delta int64, ttl time.Duration) (int64, error) {
	value, err := kv.client.HIncrBy(context.Background(), key, field, delta).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	if err := kv.refreshTTL(key, ttl); err != nil {
		return 0, err
	}
	return value, nil
}

func (kv *RedisKV) SAdd(key string, ttl time.Duration, members ...string) (int64, error) {
	if len(members) == 0 {
		return 0, errEmptyWriteValues
	}
	args := make([]any, len(members))
	for i, member := range members {
		args[i] = member
	}
	count, err := kv.client.SAdd(context.Background(), key, args...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	if len(members) > 0 {
		if err := kv.refreshTTL(key, ttl); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (kv *RedisKV) SRem(key string, members ...string) (int64, error) {
	args := make([]any, len(members))
	for i, member := range members {
		args[i] = member
	}
	count, err := kv.client.SRem(context.Background(), key, args...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return count, nil
}

func (kv *RedisKV) SContains(key string, member string) (bool, error) {
	ok, err := kv.client.SIsMember(context.Background(), key, member).Result()
	if err != nil {
		return false, normalizeRedisErr(err)
	}
	return ok, nil
}

func (kv *RedisKV) SMembers(key string) ([]string, error) {
	members, err := kv.client.SMembers(context.Background(), key).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	sort.Strings(members)
	return members, nil
}

func (kv *RedisKV) SCard(key string) (int64, error) {
	count, err := kv.client.SCard(context.Background(), key).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return count, nil
}

func (kv *RedisKV) SPop(key string, count int64) ([]string, error) {
	if count < 0 {
		return nil, errors.New("count must be non-negative")
	}
	if count == 0 {
		return []string{}, nil
	}
	values, err := kv.client.SPopN(context.Background(), key, count).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return values, nil
}

func (kv *RedisKV) LPush(key string, ttl time.Duration, values ...string) (int64, error) {
	if len(values) == 0 {
		return 0, errEmptyWriteValues
	}
	args := make([]any, len(values))
	for i, value := range values {
		args[i] = value
	}
	length, err := kv.client.LPush(context.Background(), key, args...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	if len(values) > 0 {
		if err := kv.refreshTTL(key, ttl); err != nil {
			return 0, err
		}
	}
	return length, nil
}

func (kv *RedisKV) RPush(key string, ttl time.Duration, values ...string) (int64, error) {
	if len(values) == 0 {
		return 0, errEmptyWriteValues
	}
	args := make([]any, len(values))
	for i, value := range values {
		args[i] = value
	}
	length, err := kv.client.RPush(context.Background(), key, args...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	if len(values) > 0 {
		if err := kv.refreshTTL(key, ttl); err != nil {
			return 0, err
		}
	}
	return length, nil
}

func (kv *RedisKV) LPop(key string) (string, bool, error) {
	value, err := kv.client.LPop(context.Background(), key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, normalizeRedisErr(err)
	}
	return value, true, nil
}

func (kv *RedisKV) RPop(key string) (string, bool, error) {
	value, err := kv.client.RPop(context.Background(), key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, normalizeRedisErr(err)
	}
	return value, true, nil
}

func (kv *RedisKV) LLen(key string) (int64, error) {
	length, err := kv.client.LLen(context.Background(), key).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return length, nil
}

func (kv *RedisKV) LRange(key string, start, stop int64) ([]string, error) {
	values, err := kv.client.LRange(context.Background(), key, start, stop).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return values, nil
}

func (kv *RedisKV) LTrim(key string, start, stop int64) error {
	return normalizeRedisErr(kv.client.LTrim(context.Background(), key, start, stop).Err())
}

func toRedisZMembers(members []ScoredMember) []redis.Z {
	items := make([]redis.Z, 0, len(members))
	for _, member := range members {
		items = append(items, redis.Z{
			Score:  member.Score,
			Member: member.Member,
		})
	}
	return items
}

func fromRedisZMembers(items []redis.Z) []ScoredMember {
	result := make([]ScoredMember, 0, len(items))
	for _, item := range items {
		member := ""
		switch value := item.Member.(type) {
		case string:
			member = value
		case []byte:
			member = string(value)
		default:
			member = fmt.Sprint(value)
		}
		result = append(result, ScoredMember{
			Member: member,
			Score:  item.Score,
		})
	}
	return result
}

func (kv *RedisKV) ZAdd(key string, ttl time.Duration, members ...ScoredMember) (int64, error) {
	if len(members) == 0 {
		return 0, errEmptyWriteValues
	}
	items := toRedisZMembers(members)
	count, err := kv.client.ZAdd(context.Background(), key, items...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	if len(members) > 0 {
		if err := kv.refreshTTL(key, ttl); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (kv *RedisKV) ZRem(key string, members ...string) (int64, error) {
	args := make([]any, len(members))
	for i, member := range members {
		args[i] = member
	}
	count, err := kv.client.ZRem(context.Background(), key, args...).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return count, nil
}

func (kv *RedisKV) ZScore(key string, member string) (float64, bool, error) {
	score, err := kv.client.ZScore(context.Background(), key, member).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, normalizeRedisErr(err)
	}
	return score, true, nil
}

func (kv *RedisKV) ZCard(key string) (int64, error) {
	count, err := kv.client.ZCard(context.Background(), key).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	return count, nil
}

func (kv *RedisKV) ZRange(key string, start, stop int64) ([]ScoredMember, error) {
	items, err := kv.client.ZRangeWithScores(context.Background(), key, start, stop).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return fromRedisZMembers(items), nil
}

func (kv *RedisKV) ZRevRange(key string, start, stop int64) ([]ScoredMember, error) {
	items, err := kv.client.ZRevRangeWithScores(context.Background(), key, start, stop).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return fromRedisZMembers(items), nil
}

func (kv *RedisKV) ZRangeByScore(key string, min, max float64, limit int64) ([]ScoredMember, error) {
	opt := &redis.ZRangeBy{
		Min: fmt.Sprintf("%v", min),
		Max: fmt.Sprintf("%v", max),
	}
	if limit > 0 {
		opt.Count = limit
	}
	items, err := kv.client.ZRangeByScoreWithScores(context.Background(), key, opt).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return fromRedisZMembers(items), nil
}

func (kv *RedisKV) ZPopMin(key string, count int64) ([]ScoredMember, error) {
	if count < 0 {
		return nil, errors.New("count must be non-negative")
	}
	if count == 0 {
		return []ScoredMember{}, nil
	}
	items, err := kv.client.ZPopMin(context.Background(), key, count).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return fromRedisZMembers(items), nil
}

func (kv *RedisKV) ZPopMax(key string, count int64) ([]ScoredMember, error) {
	if count < 0 {
		return nil, errors.New("count must be non-negative")
	}
	if count == 0 {
		return []ScoredMember{}, nil
	}
	items, err := kv.client.ZPopMax(context.Background(), key, count).Result()
	if err != nil {
		return nil, normalizeRedisErr(err)
	}
	return fromRedisZMembers(items), nil
}

func (kv *RedisKV) GetType() Type {
	return Redis
}

func (kv *RedisKV) Ping() error {
	return kv.client.Ping(context.Background()).Err()
}

func (kv *RedisKV) TryLock(key string, parse string, ttl time.Duration) bool {
	ok, err := kv.client.SetNX(context.Background(), key, parse, ttl).Result()
	if err = normalizeRedisErr(err); err != nil {
		kv.logger.Error(err, "error occurs when try lock: "+err.Error())
	}
	return ok
}

func (kv *RedisKV) Unlock(key string, parse string) (err error) {
	result, err := unlockScript.Run(context.Background(), kv.client, []string{key}, parse).Result()
	if err != nil {
		return normalizeRedisErr(err)
	}
	if result.(int64) == 0 {
		return errors.New("unlock failed")
	}
	return nil
}

func (kv *RedisKV) Incr(key string, delta int64, ttl time.Duration, createTTLOnly bool) (int64, error) {
	result, err := kv.client.IncrBy(context.Background(), key, delta).Result()
	if err != nil {
		return 0, normalizeRedisErr(err)
	}
	if ttl > 0 {
		if createTTLOnly {
			err = kv.client.ExpireNX(context.Background(), key, ttl).Err()
		} else {
			err = kv.client.Expire(context.Background(), key, ttl).Err()
		}
		if err != nil {
			return 0, normalizeRedisErr(err)
		}
	}
	return result, nil
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
		return false, normalizeRedisErr(err)
	}
	ok := result == 1
	if !ok {
		lock.locked = false
	}
	return ok, nil
}
