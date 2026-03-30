/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:26
 */

// Package kv

package kv

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/cache"
	"github.com/real-uangi/allingo/common/log"
)

type localValueKind uint8

const (
	localStringValue localValueKind = iota + 1
	localHashValue
	localSetValue
	localListValue
	localSortedSetValue
)

type localEntry struct {
	kind       localValueKind
	expiration int64
	data       any
}

func (entry *localEntry) expired(now int64) bool {
	return entry.expiration > 0 && entry.expiration <= now
}

type LocalKV struct {
	mu     sync.Mutex
	data   map[string]*localEntry
	locks  *cache.Cache[string]
	logger *log.StdLogger
}

func newLocalKV() *LocalKV {
	logger := log.For[LocalKV]()
	logger.Warn("using local kv, standalone mode only")
	return &LocalKV{
		data:   make(map[string]*localEntry),
		locks:  cache.New[string](time.Minute),
		logger: logger,
	}
}

func localExpireAt(ttl time.Duration, now int64) int64 {
	if ttl <= 0 {
		return 0
	}
	return now + ttl.Milliseconds()
}

func (kv *LocalKV) loadEntryLocked(key string, now int64) (*localEntry, bool) {
	entry, ok := kv.data[key]
	if !ok {
		return nil, false
	}
	if entry.expired(now) {
		delete(kv.data, key)
		return nil, false
	}
	return entry, true
}

func (kv *LocalKV) newEntry(kind localValueKind, data any, ttl time.Duration, now int64) *localEntry {
	return &localEntry{
		kind:       kind,
		expiration: localExpireAt(ttl, now),
		data:       data,
	}
}

func (kv *LocalKV) setEntryLocked(key string, kind localValueKind, data any, ttl time.Duration, now int64) {
	kv.data[key] = kv.newEntry(kind, data, ttl, now)
}

func (kv *LocalKV) refreshEntryTTLLocked(entry *localEntry, ttl time.Duration, now int64) {
	if ttl > 0 {
		entry.expiration = localExpireAt(ttl, now)
	}
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return make(map[string]string)
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneStringSet(src map[string]struct{}) map[string]struct{} {
	if len(src) == 0 {
		return make(map[string]struct{})
	}
	dst := make(map[string]struct{}, len(src))
	for member := range src {
		dst[member] = struct{}{}
	}
	return dst
}

func cloneStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func cloneScoredMemberMap(src map[string]float64) map[string]float64 {
	if len(src) == 0 {
		return make(map[string]float64)
	}
	dst := make(map[string]float64, len(src))
	for member, score := range src {
		dst[member] = score
	}
	return dst
}

func normalizeRange(start, stop int64, length int) (int, int, bool) {
	if length == 0 {
		return 0, 0, false
	}
	if start < 0 {
		start = int64(length) + start
	}
	if stop < 0 {
		stop = int64(length) + stop
	}
	if start < 0 {
		start = 0
	}
	if stop < 0 {
		return 0, 0, false
	}
	if start >= int64(length) {
		return 0, 0, false
	}
	if stop >= int64(length) {
		stop = int64(length - 1)
	}
	if start > stop {
		return 0, 0, false
	}
	return int(start), int(stop), true
}

func sortScoredMembersAsc(members map[string]float64) []ScoredMember {
	items := make([]ScoredMember, 0, len(members))
	for member, score := range members {
		items = append(items, ScoredMember{Member: member, Score: score})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].Member < items[j].Member
		}
		return items[i].Score < items[j].Score
	})
	return items
}

func reverseScoredMembers(items []ScoredMember) []ScoredMember {
	reversed := cloneScoredMembers(items)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}

func cloneScoredMembers(src []ScoredMember) []ScoredMember {
	if len(src) == 0 {
		return nil
	}
	dst := make([]ScoredMember, len(src))
	copy(dst, src)
	return dst
}

func (kv *LocalKV) entryKindError(entry *localEntry, expected localValueKind) error {
	if entry == nil || entry.kind == expected {
		return nil
	}
	return ErrWrongType
}

func (kv *LocalKV) Exists(key string) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	_, ok := kv.loadEntryLocked(key, now)
	return ok, nil
}

func (kv *LocalKV) Del(key string) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if _, ok := kv.loadEntryLocked(key, now); !ok {
		return false, nil
	}
	delete(kv.data, key)
	return true, nil
}

func (kv *LocalKV) Expire(key string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, nil
	}
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return false, nil
	}
	entry.expiration = localExpireAt(ttl, now)
	return true, nil
}

func (kv *LocalKV) Persist(key string) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok || entry.expiration == 0 {
		return false, nil
	}
	entry.expiration = 0
	return true, nil
}

func (kv *LocalKV) TTL(key string) (time.Duration, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, false, nil
	}
	if entry.expiration == 0 {
		return 0, true, nil
	}
	ttl := time.Duration(entry.expiration-now) * time.Millisecond
	if ttl < 0 {
		ttl = 0
	}
	return ttl, true, nil
}

func (kv *LocalKV) Set(key string, value string, ttl time.Duration) error {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if entry, ok := kv.loadEntryLocked(key, now); ok && entry.kind != localStringValue {
		return ErrWrongType
	}
	kv.setEntryLocked(key, localStringValue, value, ttl, now)
	return nil
}

func (kv *LocalKV) Get(key string) (string, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return "", false, nil
	}
	if err := kv.entryKindError(entry, localStringValue); err != nil {
		return "", false, err
	}
	return entry.data.(string), true, nil
}

func (kv *LocalKV) SetIfAbsent(key string, value string, ttl time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if ok {
		if err := kv.entryKindError(entry, localStringValue); err != nil {
			return false, err
		}
		return false, nil
	}
	kv.setEntryLocked(key, localStringValue, value, ttl, now)
	return true, nil
}

func (kv *LocalKV) SetIfPresent(key string, value string, ttl time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return false, nil
	}
	if err := kv.entryKindError(entry, localStringValue); err != nil {
		return false, err
	}
	entry.data = value
	kv.refreshEntryTTLLocked(entry, ttl, now)
	return true, nil
}

func (kv *LocalKV) CompareAndSet(key string, expected, value string, ttl time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return false, nil
	}
	if err := kv.entryKindError(entry, localStringValue); err != nil {
		return false, err
	}
	if entry.data.(string) != expected {
		return false, nil
	}
	entry.data = value
	kv.refreshEntryTTLLocked(entry, ttl, now)
	return true, nil
}

func (kv *LocalKV) CompareAndDelete(key string, expected string) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return false, nil
	}
	if err := kv.entryKindError(entry, localStringValue); err != nil {
		return false, err
	}
	if entry.data.(string) != expected {
		return false, nil
	}
	delete(kv.data, key)
	return true, nil
}

func (kv *LocalKV) GetAndDelete(key string) (string, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return "", false, nil
	}
	if err := kv.entryKindError(entry, localStringValue); err != nil {
		return "", false, err
	}
	delete(kv.data, key)
	return entry.data.(string), true, nil
}

func (kv *LocalKV) GetAndSet(key string, value string, ttl time.Duration) (string, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		kv.setEntryLocked(key, localStringValue, value, ttl, now)
		return "", false, nil
	}
	if err := kv.entryKindError(entry, localStringValue); err != nil {
		return "", false, err
	}
	old := entry.data.(string)
	entry.data = value
	kv.refreshEntryTTLLocked(entry, ttl, now)
	return old, true, nil
}

func (kv *LocalKV) MGet(keys ...string) (map[string]string, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		entry, ok := kv.loadEntryLocked(key, now)
		if !ok {
			continue
		}
		if err := kv.entryKindError(entry, localStringValue); err != nil {
			return nil, err
		}
		result[key] = entry.data.(string)
	}
	return result, nil
}

func (kv *LocalKV) MSet(values map[string]string, ttl time.Duration) error {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	for key := range values {
		entry, ok := kv.loadEntryLocked(key, now)
		if ok && entry.kind != localStringValue {
			return ErrWrongType
		}
	}
	for key, value := range values {
		kv.setEntryLocked(key, localStringValue, value, ttl, now)
	}
	return nil
}

func (kv *LocalKV) MDel(keys ...string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	var deleted int64
	for _, key := range keys {
		if _, ok := kv.loadEntryLocked(key, now); ok {
			delete(kv.data, key)
			deleted++
		}
	}
	return deleted, nil
}

func (kv *LocalKV) SetStruct(key string, obj any, ttl time.Duration) error {
	str, err := anyToString(obj)
	if err != nil {
		return err
	}
	return kv.Set(key, str, ttl)
}

func (kv *LocalKV) GetStruct(key string, p any) (bool, error) {
	str, ok, err := kv.Get(key)
	if err != nil || !ok {
		return ok, err
	}
	return true, stringToAny(str, p)
}

func (kv *LocalKV) HSet(key string, field string, value string, ttl time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		hash := map[string]string{field: value}
		kv.setEntryLocked(key, localHashValue, hash, ttl, now)
		return true, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return false, err
	}
	hash := cloneStringMap(entry.data.(map[string]string))
	_, added := hash[field]
	hash[field] = value
	entry.data = hash
	kv.refreshEntryTTLLocked(entry, ttl, now)
	return !added, nil
}

func (kv *LocalKV) HSetIfAbsent(key string, field string, value string, ttl time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		hash := map[string]string{field: value}
		kv.setEntryLocked(key, localHashValue, hash, ttl, now)
		return true, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return false, err
	}
	hash := cloneStringMap(entry.data.(map[string]string))
	if _, exists := hash[field]; exists {
		kv.refreshEntryTTLLocked(entry, ttl, now)
		return false, nil
	}
	hash[field] = value
	entry.data = hash
	kv.refreshEntryTTLLocked(entry, ttl, now)
	return true, nil
}

func (kv *LocalKV) HGet(key string, field string) (string, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return "", false, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return "", false, err
	}
	value, exists := entry.data.(map[string]string)[field]
	return value, exists, nil
}

func (kv *LocalKV) HDel(key string, fields ...string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return 0, err
	}
	hash := cloneStringMap(entry.data.(map[string]string))
	var deleted int64
	for _, field := range fields {
		if _, exists := hash[field]; exists {
			delete(hash, field)
			deleted++
		}
	}
	if len(hash) == 0 {
		delete(kv.data, key)
		return deleted, nil
	}
	entry.data = hash
	return deleted, nil
}

func (kv *LocalKV) HExists(key string, field string) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return false, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return false, err
	}
	_, exists := entry.data.(map[string]string)[field]
	return exists, nil
}

func (kv *LocalKV) HGetAll(key string) (map[string]string, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return map[string]string{}, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return nil, err
	}
	return cloneStringMap(entry.data.(map[string]string)), nil
}

func (kv *LocalKV) HLen(key string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return 0, err
	}
	return int64(len(entry.data.(map[string]string))), nil
}

func (kv *LocalKV) HIncr(key string, field string, delta int64, ttl time.Duration) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		hash := map[string]string{field: strconv.FormatInt(delta, 10)}
		kv.setEntryLocked(key, localHashValue, hash, ttl, now)
		return delta, nil
	}
	if err := kv.entryKindError(entry, localHashValue); err != nil {
		return 0, err
	}
	hash := cloneStringMap(entry.data.(map[string]string))
	current := int64(0)
	if value, exists := hash[field]; exists {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, err
		}
		current = parsed
	}
	current += delta
	hash[field] = strconv.FormatInt(current, 10)
	entry.data = hash
	kv.refreshEntryTTLLocked(entry, ttl, now)
	return current, nil
}

func (kv *LocalKV) SAdd(key string, ttl time.Duration, members ...string) (int64, error) {
	if len(members) == 0 {
		return 0, errEmptyWriteValues
	}
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		set := make(map[string]struct{}, len(members))
		var added int64
		for _, member := range members {
			if _, exists := set[member]; !exists {
				set[member] = struct{}{}
				added++
			}
		}
		kv.setEntryLocked(key, localSetValue, set, ttl, now)
		return added, nil
	}
	if err := kv.entryKindError(entry, localSetValue); err != nil {
		return 0, err
	}
	set := cloneStringSet(entry.data.(map[string]struct{}))
	var added int64
	for _, member := range members {
		if _, exists := set[member]; !exists {
			set[member] = struct{}{}
			added++
		}
	}
	entry.data = set
	kv.refreshEntryTTLLocked(entry, ttl, now)
	return added, nil
}

func (kv *LocalKV) SRem(key string, members ...string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, nil
	}
	if err := kv.entryKindError(entry, localSetValue); err != nil {
		return 0, err
	}
	set := cloneStringSet(entry.data.(map[string]struct{}))
	var removed int64
	for _, member := range members {
		if _, exists := set[member]; exists {
			delete(set, member)
			removed++
		}
	}
	if len(set) == 0 {
		delete(kv.data, key)
		return removed, nil
	}
	entry.data = set
	return removed, nil
}

func (kv *LocalKV) SContains(key string, member string) (bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return false, nil
	}
	if err := kv.entryKindError(entry, localSetValue); err != nil {
		return false, err
	}
	_, exists := entry.data.(map[string]struct{})[member]
	return exists, nil
}

func (kv *LocalKV) SMembers(key string) ([]string, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return []string{}, nil
	}
	if err := kv.entryKindError(entry, localSetValue); err != nil {
		return nil, err
	}
	set := entry.data.(map[string]struct{})
	members := make([]string, 0, len(set))
	for member := range set {
		members = append(members, member)
	}
	sort.Strings(members)
	return members, nil
}

func (kv *LocalKV) SCard(key string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, nil
	}
	if err := kv.entryKindError(entry, localSetValue); err != nil {
		return 0, err
	}
	return int64(len(entry.data.(map[string]struct{}))), nil
}

func (kv *LocalKV) SPop(key string, count int64) ([]string, error) {
	if count < 0 {
		return nil, errors.New("count must be non-negative")
	}
	if count == 0 {
		return []string{}, nil
	}
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return []string{}, nil
	}
	if err := kv.entryKindError(entry, localSetValue); err != nil {
		return nil, err
	}
	set := cloneStringSet(entry.data.(map[string]struct{}))
	capacity := len(set)
	if int(count) < capacity {
		capacity = int(count)
	}
	popped := make([]string, 0, capacity)
	for member := range set {
		popped = append(popped, member)
		if int64(len(popped)) == count {
			break
		}
	}
	for _, member := range popped {
		delete(set, member)
	}
	if len(set) == 0 {
		delete(kv.data, key)
	} else {
		entry.data = set
	}
	return popped, nil
}

func (kv *LocalKV) LPush(key string, ttl time.Duration, values ...string) (int64, error) {
	if len(values) == 0 {
		return 0, errEmptyWriteValues
	}
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	current := []string(nil)
	if ok {
		if err := kv.entryKindError(entry, localListValue); err != nil {
			return 0, err
		}
		current = cloneStringSlice(entry.data.([]string))
	}
	prefix := cloneStringSlice(values)
	for left, right := 0, len(prefix)-1; left < right; left, right = left+1, right-1 {
		prefix[left], prefix[right] = prefix[right], prefix[left]
	}
	current = append(prefix, current...)
	if ok {
		entry.data = current
		kv.refreshEntryTTLLocked(entry, ttl, now)
		return int64(len(current)), nil
	}
	kv.setEntryLocked(key, localListValue, current, ttl, now)
	return int64(len(current)), nil
}

func (kv *LocalKV) RPush(key string, ttl time.Duration, values ...string) (int64, error) {
	if len(values) == 0 {
		return 0, errEmptyWriteValues
	}
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	current := []string(nil)
	if ok {
		if err := kv.entryKindError(entry, localListValue); err != nil {
			return 0, err
		}
		current = cloneStringSlice(entry.data.([]string))
	}
	current = append(current, values...)
	if ok {
		entry.data = current
		kv.refreshEntryTTLLocked(entry, ttl, now)
		return int64(len(current)), nil
	}
	kv.setEntryLocked(key, localListValue, current, ttl, now)
	return int64(len(current)), nil
}

func (kv *LocalKV) LPop(key string) (string, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return "", false, nil
	}
	if err := kv.entryKindError(entry, localListValue); err != nil {
		return "", false, err
	}
	list := cloneStringSlice(entry.data.([]string))
	if len(list) == 0 {
		delete(kv.data, key)
		return "", false, nil
	}
	value := list[0]
	list = list[1:]
	if len(list) == 0 {
		delete(kv.data, key)
	} else {
		entry.data = list
	}
	return value, true, nil
}

func (kv *LocalKV) RPop(key string) (string, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return "", false, nil
	}
	if err := kv.entryKindError(entry, localListValue); err != nil {
		return "", false, err
	}
	list := cloneStringSlice(entry.data.([]string))
	if len(list) == 0 {
		delete(kv.data, key)
		return "", false, nil
	}
	last := len(list) - 1
	value := list[last]
	list = list[:last]
	if len(list) == 0 {
		delete(kv.data, key)
	} else {
		entry.data = list
	}
	return value, true, nil
}

func (kv *LocalKV) LLen(key string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, nil
	}
	if err := kv.entryKindError(entry, localListValue); err != nil {
		return 0, err
	}
	return int64(len(entry.data.([]string))), nil
}

func (kv *LocalKV) LRange(key string, start, stop int64) ([]string, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return []string{}, nil
	}
	if err := kv.entryKindError(entry, localListValue); err != nil {
		return nil, err
	}
	list := entry.data.([]string)
	from, to, ok := normalizeRange(start, stop, len(list))
	if !ok {
		return []string{}, nil
	}
	return cloneStringSlice(list[from : to+1]), nil
}

func (kv *LocalKV) LTrim(key string, start, stop int64) error {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return nil
	}
	if err := kv.entryKindError(entry, localListValue); err != nil {
		return err
	}
	list := entry.data.([]string)
	from, to, ok := normalizeRange(start, stop, len(list))
	if !ok {
		delete(kv.data, key)
		return nil
	}
	entry.data = cloneStringSlice(list[from : to+1])
	return nil
}

func (kv *LocalKV) ZAdd(key string, ttl time.Duration, members ...ScoredMember) (int64, error) {
	if len(members) == 0 {
		return 0, errEmptyWriteValues
	}
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	current := map[string]float64{}
	if ok {
		if err := kv.entryKindError(entry, localSortedSetValue); err != nil {
			return 0, err
		}
		current = cloneScoredMemberMap(entry.data.(map[string]float64))
	}
	var added int64
	for _, member := range members {
		if _, exists := current[member.Member]; !exists {
			added++
		}
		current[member.Member] = member.Score
	}
	if ok {
		entry.data = current
		kv.refreshEntryTTLLocked(entry, ttl, now)
		return added, nil
	}
	kv.setEntryLocked(key, localSortedSetValue, current, ttl, now)
	return added, nil
}

func (kv *LocalKV) ZRem(key string, members ...string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, nil
	}
	if err := kv.entryKindError(entry, localSortedSetValue); err != nil {
		return 0, err
	}
	current := cloneScoredMemberMap(entry.data.(map[string]float64))
	var removed int64
	for _, member := range members {
		if _, exists := current[member]; exists {
			delete(current, member)
			removed++
		}
	}
	if len(current) == 0 {
		delete(kv.data, key)
		return removed, nil
	}
	entry.data = current
	return removed, nil
}

func (kv *LocalKV) ZScore(key string, member string) (float64, bool, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, false, nil
	}
	if err := kv.entryKindError(entry, localSortedSetValue); err != nil {
		return 0, false, err
	}
	score, exists := entry.data.(map[string]float64)[member]
	return score, exists, nil
}

func (kv *LocalKV) ZCard(key string) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return 0, nil
	}
	if err := kv.entryKindError(entry, localSortedSetValue); err != nil {
		return 0, err
	}
	return int64(len(entry.data.(map[string]float64))), nil
}

func (kv *LocalKV) ZRange(key string, start, stop int64) ([]ScoredMember, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return []ScoredMember{}, nil
	}
	if err := kv.entryKindError(entry, localSortedSetValue); err != nil {
		return nil, err
	}
	items := sortScoredMembersAsc(entry.data.(map[string]float64))
	from, to, ok := normalizeRange(start, stop, len(items))
	if !ok {
		return []ScoredMember{}, nil
	}
	return cloneScoredMembers(items[from : to+1]), nil
}

func (kv *LocalKV) ZRevRange(key string, start, stop int64) ([]ScoredMember, error) {
	items, err := kv.ZRange(key, 0, -1)
	if err != nil {
		return nil, err
	}
	items = reverseScoredMembers(items)
	from, to, ok := normalizeRange(start, stop, len(items))
	if !ok {
		return []ScoredMember{}, nil
	}
	return cloneScoredMembers(items[from : to+1]), nil
}

func (kv *LocalKV) ZRangeByScore(key string, min, max float64, limit int64) ([]ScoredMember, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return []ScoredMember{}, nil
	}
	if err := kv.entryKindError(entry, localSortedSetValue); err != nil {
		return nil, err
	}
	items := sortScoredMembersAsc(entry.data.(map[string]float64))
	filtered := make([]ScoredMember, 0, len(items))
	for _, item := range items {
		if item.Score < min || item.Score > max {
			continue
		}
		filtered = append(filtered, item)
		if limit > 0 && int64(len(filtered)) >= limit {
			break
		}
	}
	return cloneScoredMembers(filtered), nil
}

func (kv *LocalKV) ZPopMin(key string, count int64) ([]ScoredMember, error) {
	return kv.zPop(key, count, false)
}

func (kv *LocalKV) ZPopMax(key string, count int64) ([]ScoredMember, error) {
	return kv.zPop(key, count, true)
}

func (kv *LocalKV) zPop(key string, count int64, max bool) ([]ScoredMember, error) {
	if count < 0 {
		return nil, errors.New("count must be non-negative")
	}
	if count == 0 {
		return []ScoredMember{}, nil
	}
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		return []ScoredMember{}, nil
	}
	if err := kv.entryKindError(entry, localSortedSetValue); err != nil {
		return nil, err
	}
	items := sortScoredMembersAsc(entry.data.(map[string]float64))
	if max {
		items = reverseScoredMembers(items)
	}
	if int(count) > len(items) {
		count = int64(len(items))
	}
	popped := cloneScoredMembers(items[:count])
	current := cloneScoredMemberMap(entry.data.(map[string]float64))
	for _, item := range popped {
		delete(current, item.Member)
	}
	if len(current) == 0 {
		delete(kv.data, key)
	} else {
		entry.data = current
	}
	return popped, nil
}

func (kv *LocalKV) GetType() Type {
	return Local
}

func (kv *LocalKV) Ping() error {
	return nil
}

func (kv *LocalKV) Incr(key string, delta int64, ttl time.Duration, createTTLOnly bool) (int64, error) {
	now := time.Now().UnixMilli()
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.loadEntryLocked(key, now)
	if !ok {
		kv.setEntryLocked(key, localStringValue, strconv.FormatInt(delta, 10), ttl, now)
		return delta, nil
	}
	if err := kv.entryKindError(entry, localStringValue); err != nil {
		return 0, err
	}
	current, err := strconv.ParseInt(entry.data.(string), 10, 64)
	if err != nil {
		return 0, err
	}
	current += delta
	entry.data = strconv.FormatInt(current, 10)
	if ttl > 0 {
		if createTTLOnly {
			if entry.expiration == 0 {
				entry.expiration = localExpireAt(ttl, now)
			}
		} else {
			entry.expiration = localExpireAt(ttl, now)
		}
	}
	return current, nil
}

type LocalLock struct {
	kv     *LocalKV
	key    string
	parse  string
	locked bool
}

func (kv *LocalKV) NewLock(key string) Lock {
	return &LocalLock{
		kv:     kv,
		key:    key,
		parse:  uuid.NewString(),
		locked: false,
	}
}

func (lock *LocalLock) TryLock(ttl time.Duration) bool {
	lock.locked = lock.kv.locks.TryLock(lock.key, lock.parse, ttl)
	return lock.locked
}

func (lock *LocalLock) Unlock() error {
	if !lock.locked {
		return nil
	}
	err := lock.kv.locks.Unlock(lock.key, lock.parse)
	if err == nil {
		lock.locked = false
	}
	return err
}

func (lock *LocalLock) Lock(ttl, maxWait time.Duration) error {
	err := lock.kv.locks.Lock(lock.key, lock.parse, ttl, maxWait)
	if err != nil {
		return err
	}
	lock.locked = true
	return nil
}

func (lock *LocalLock) Refresh(ttl time.Duration) (bool, error) {
	if !lock.locked {
		return false, nil
	}
	ok := lock.kv.locks.RefreshLock(lock.key, lock.parse, ttl)
	if !ok {
		lock.locked = false
	}
	return ok, nil
}
