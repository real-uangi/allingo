package kv

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type kvFactory struct {
	name   string
	create func(t *testing.T) KV
}

func TestLocalKVContract(t *testing.T) {
	runKVContractTests(t, kvFactory{
		name: "local",
		create: func(t *testing.T) KV {
			t.Helper()
			return newLocalKV()
		},
	})
}

func TestRedisKVContract(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is not set")
	}
	password := os.Getenv("REDIS_PASSWORD")
	store := newRedisKV(addr, password)
	if err := store.Ping(); err != nil {
		t.Skipf("redis is not reachable: %v", err)
	}
	runKVContractTests(t, kvFactory{
		name: "redis",
		create: func(t *testing.T) KV {
			t.Helper()
			return store
		},
	})
}

func runKVContractTests(t *testing.T, factory kvFactory) {
	t.Helper()
	store := factory.create(t)

	t.Run(factory.name+"/string_and_ttl", func(t *testing.T) {
		key := testKey(factory.name, "string")
		exists, err := store.Exists(key)
		if err != nil {
			t.Fatalf("Exists returned error: %v", err)
		}
		if exists {
			t.Fatal("expected key to be missing before Set")
		}

		if err := store.Set(key, "value-1", 120*time.Millisecond); err != nil {
			t.Fatalf("Set returned error: %v", err)
		}
		value, ok, err := store.Get(key)
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if !ok || value != "value-1" {
			t.Fatalf("expected value-1, got ok=%v value=%q", ok, value)
		}

		ttl, ok, err := store.TTL(key)
		if err != nil {
			t.Fatalf("TTL returned error: %v", err)
		}
		if !ok || ttl <= 0 {
			t.Fatalf("expected positive TTL, got ok=%v ttl=%v", ok, ttl)
		}

		ok, err = store.Persist(key)
		if err != nil {
			t.Fatalf("Persist returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected Persist to remove TTL")
		}

		ttl, ok, err = store.TTL(key)
		if err != nil {
			t.Fatalf("TTL after Persist returned error: %v", err)
		}
		if !ok || ttl != 0 {
			t.Fatalf("expected persistent key TTL to be 0, got ok=%v ttl=%v", ok, ttl)
		}

		ok, err = store.Expire(key, 80*time.Millisecond)
		if err != nil {
			t.Fatalf("Expire returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected Expire to succeed")
		}

		type payload struct {
			Name string `json:"name"`
		}

		structKey := testKey(factory.name, "struct")
		if err := store.SetStruct(structKey, payload{Name: "xu"}, time.Second); err != nil {
			t.Fatalf("SetStruct returned error: %v", err)
		}
		var got payload
		ok, err = store.GetStruct(structKey, &got)
		if err != nil {
			t.Fatalf("GetStruct returned error: %v", err)
		}
		if !ok || got.Name != "xu" {
			t.Fatalf("expected GetStruct to decode payload, got ok=%v payload=%+v", ok, got)
		}

		time.Sleep(100 * time.Millisecond)
		exists, err = store.Exists(key)
		if err != nil {
			t.Fatalf("Exists after expiration returned error: %v", err)
		}
		if exists {
			t.Fatal("expected key to expire")
		}
	})

	t.Run(factory.name+"/atomic_string_ops", func(t *testing.T) {
		key := testKey(factory.name, "atomic")

		ok, err := store.SetIfPresent(key, "missing", time.Second)
		if err != nil {
			t.Fatalf("SetIfPresent on missing key returned error: %v", err)
		}
		if ok {
			t.Fatal("expected SetIfPresent to fail for missing key")
		}

		ok, err = store.SetIfAbsent(key, "value-1", time.Second)
		if err != nil {
			t.Fatalf("SetIfAbsent returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected first SetIfAbsent to succeed")
		}

		ok, err = store.SetIfAbsent(key, "value-2", time.Second)
		if err != nil {
			t.Fatalf("second SetIfAbsent returned error: %v", err)
		}
		if ok {
			t.Fatal("expected second SetIfAbsent to fail")
		}

		ok, err = store.SetIfPresent(key, "value-2", 120*time.Millisecond)
		if err != nil {
			t.Fatalf("SetIfPresent returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected SetIfPresent to succeed")
		}

		ok, err = store.CompareAndSet(key, "wrong", "value-3", 0)
		if err != nil {
			t.Fatalf("CompareAndSet with wrong value returned error: %v", err)
		}
		if ok {
			t.Fatal("expected CompareAndSet to fail when expected value mismatches")
		}

		ok, err = store.CompareAndSet(key, "value-2", "value-3", 0)
		if err != nil {
			t.Fatalf("CompareAndSet returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected CompareAndSet to succeed")
		}

		old, existed, err := store.GetAndSet(key, "value-4", 0)
		if err != nil {
			t.Fatalf("GetAndSet returned error: %v", err)
		}
		if !existed || old != "value-3" {
			t.Fatalf("expected GetAndSet to return value-3, got existed=%v old=%q", existed, old)
		}

		value, ok, err := store.Get(key)
		if err != nil {
			t.Fatalf("Get after GetAndSet returned error: %v", err)
		}
		if !ok || value != "value-4" {
			t.Fatalf("expected key to contain value-4, got ok=%v value=%q", ok, value)
		}

		old, existed, err = store.GetAndDelete(key)
		if err != nil {
			t.Fatalf("GetAndDelete returned error: %v", err)
		}
		if !existed || old != "value-4" {
			t.Fatalf("expected GetAndDelete to return value-4, got existed=%v old=%q", existed, old)
		}

		value, ok, err = store.Get(key)
		if err != nil {
			t.Fatalf("Get after GetAndDelete returned error: %v", err)
		}
		if ok || value != "" {
			t.Fatalf("expected key to be deleted, got ok=%v value=%q", ok, value)
		}

		missingKey := testKey(factory.name, "atomic-missing")
		old, existed, err = store.GetAndSet(missingKey, "created", 0)
		if err != nil {
			t.Fatalf("GetAndSet on missing key returned error: %v", err)
		}
		if existed || old != "" {
			t.Fatalf("expected GetAndSet on missing key to report existed=false, got existed=%v old=%q", existed, old)
		}

		hashKey := testKey(factory.name, "atomic-wrong-type")
		if _, err := store.HSet(hashKey, "field", "value", time.Second); err != nil {
			t.Fatalf("HSet for wrong-type test returned error: %v", err)
		}
		if _, _, err := store.Get(hashKey); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected Get on hash key to return ErrWrongType, got %v", err)
		}
		if err := store.Set(hashKey, "string", time.Second); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected Set on hash key to return ErrWrongType, got %v", err)
		}
		if _, err := store.SetIfAbsent(hashKey, "string", time.Second); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected SetIfAbsent on hash key to return ErrWrongType, got %v", err)
		}
	})

	t.Run(factory.name+"/batch_ops", func(t *testing.T) {
		first := testKey(factory.name, "mset-1")
		second := testKey(factory.name, "mset-2")
		if err := store.MSet(map[string]string{
			first:  "a",
			second: "b",
		}, 90*time.Millisecond); err != nil {
			t.Fatalf("MSet returned error: %v", err)
		}

		values, err := store.MGet(first, second, testKey(factory.name, "missing"))
		if err != nil {
			t.Fatalf("MGet returned error: %v", err)
		}
		if len(values) != 2 || values[first] != "a" || values[second] != "b" {
			t.Fatalf("unexpected MGet result: %+v", values)
		}

		time.Sleep(110 * time.Millisecond)
		values, err = store.MGet(first, second)
		if err != nil {
			t.Fatalf("MGet after expiration returned error: %v", err)
		}
		if len(values) != 0 {
			t.Fatalf("expected MGet to return no values after expiration, got %+v", values)
		}

		hashKey := testKey(factory.name, "mset-wrong-type")
		stringKey := testKey(factory.name, "mset-good-key")
		if _, err := store.HSet(hashKey, "field", "value", time.Second); err != nil {
			t.Fatalf("HSet for MSet wrong-type test returned error: %v", err)
		}
		if err := store.MSet(map[string]string{
			hashKey:   "x",
			stringKey: "y",
		}, time.Second); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected MSet to return ErrWrongType, got %v", err)
		}
		if _, ok, err := store.Get(stringKey); err != nil || ok {
			t.Fatalf("expected MSet failure to avoid writing string key, got ok=%v err=%v", ok, err)
		}

		if err := store.MSet(map[string]string{
			first:  "1",
			second: "2",
		}, 0); err != nil {
			t.Fatalf("second MSet returned error: %v", err)
		}
		deleted, err := store.MDel(first, second, testKey(factory.name, "missing-2"))
		if err != nil {
			t.Fatalf("MDel returned error: %v", err)
		}
		if deleted != 2 {
			t.Fatalf("expected MDel to delete 2 keys, got %d", deleted)
		}
	})

	t.Run(factory.name+"/counter_ops", func(t *testing.T) {
		key := testKey(factory.name, "counter")
		value, err := store.Incr(key, 1, 120*time.Millisecond, true)
		if err != nil {
			t.Fatalf("first Incr returned error: %v", err)
		}
		if value != 1 {
			t.Fatalf("expected first Incr to return 1, got %d", value)
		}

		raw, ok, err := store.Get(key)
		if err != nil {
			t.Fatalf("Get after first Incr returned error: %v", err)
		}
		if !ok || raw != "1" {
			t.Fatalf("expected string counter value 1, got ok=%v raw=%q", ok, raw)
		}

		time.Sleep(70 * time.Millisecond)
		value, err = store.Incr(key, 1, 120*time.Millisecond, true)
		if err != nil {
			t.Fatalf("second Incr returned error: %v", err)
		}
		if value != 2 {
			t.Fatalf("expected second Incr to return 2, got %d", value)
		}

		time.Sleep(70 * time.Millisecond)
		exists, err := store.Exists(key)
		if err != nil {
			t.Fatalf("Exists after non-refreshed Incr returned error: %v", err)
		}
		if exists {
			t.Fatal("expected counter to expire when createTTLOnly=true")
		}

		value, err = store.Incr(key, 1, 120*time.Millisecond, false)
		if err != nil {
			t.Fatalf("third Incr returned error: %v", err)
		}
		if value != 1 {
			t.Fatalf("expected recreated counter to return 1, got %d", value)
		}
		time.Sleep(70 * time.Millisecond)
		value, err = store.Incr(key, 1, 120*time.Millisecond, false)
		if err != nil {
			t.Fatalf("fourth Incr returned error: %v", err)
		}
		if value != 2 {
			t.Fatalf("expected refreshed counter to return 2, got %d", value)
		}
		time.Sleep(70 * time.Millisecond)
		exists, err = store.Exists(key)
		if err != nil {
			t.Fatalf("Exists after refreshed Incr returned error: %v", err)
		}
		if !exists {
			t.Fatal("expected counter TTL to be refreshed when createTTLOnly=false")
		}

		persistentKey := testKey(factory.name, "counter-persistent")
		if err := store.Set(persistentKey, "9", 0); err != nil {
			t.Fatalf("Set for persistent counter returned error: %v", err)
		}
		value, err = store.Incr(persistentKey, 1, 120*time.Millisecond, true)
		if err != nil {
			t.Fatalf("Incr on persistent counter returned error: %v", err)
		}
		if value != 10 {
			t.Fatalf("expected persistent counter to return 10, got %d", value)
		}
		ttl, ok, err := store.TTL(persistentKey)
		if err != nil {
			t.Fatalf("TTL on persistent counter returned error: %v", err)
		}
		if !ok || ttl <= 0 {
			t.Fatalf("expected createTTLOnly=true to attach TTL to persistent key, got ok=%v ttl=%v", ok, ttl)
		}

		hashKey := testKey(factory.name, "counter-wrong-type")
		if _, err := store.HSet(hashKey, "field", "value", time.Second); err != nil {
			t.Fatalf("HSet for counter wrong-type test returned error: %v", err)
		}
		if _, err := store.Incr(hashKey, 1, time.Second, false); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected Incr on hash key to return ErrWrongType, got %v", err)
		}
	})

	t.Run(factory.name+"/hash_ops", func(t *testing.T) {
		key := testKey(factory.name, "hash")
		added, err := store.HSet(key, "name", "xu", 120*time.Millisecond)
		if err != nil {
			t.Fatalf("HSet returned error: %v", err)
		}
		if !added {
			t.Fatal("expected first HSet to add a field")
		}

		added, err = store.HSet(key, "name", "boss", 0)
		if err != nil {
			t.Fatalf("second HSet returned error: %v", err)
		}
		if added {
			t.Fatal("expected second HSet to update existing field")
		}

		added, err = store.HSetIfAbsent(key, "name", "ignored", 0)
		if err != nil {
			t.Fatalf("HSetIfAbsent on existing field returned error: %v", err)
		}
		if added {
			t.Fatal("expected HSetIfAbsent on existing field to fail")
		}

		added, err = store.HSetIfAbsent(key, "role", "admin", 0)
		if err != nil {
			t.Fatalf("HSetIfAbsent returned error: %v", err)
		}
		if !added {
			t.Fatal("expected HSetIfAbsent to add a new field")
		}

		refreshKey := testKey(factory.name, "hash-refresh")
		added, err = store.HSet(refreshKey, "name", "xu", 80*time.Millisecond)
		if err != nil {
			t.Fatalf("HSet for TTL refresh test returned error: %v", err)
		}
		if !added {
			t.Fatal("expected HSet for TTL refresh test to add a field")
		}
		time.Sleep(50 * time.Millisecond)
		added, err = store.HSetIfAbsent(refreshKey, "name", "ignored", 120*time.Millisecond)
		if err != nil {
			t.Fatalf("HSetIfAbsent for TTL refresh test returned error: %v", err)
		}
		if added {
			t.Fatal("expected HSetIfAbsent to report existing field in TTL refresh test")
		}
		time.Sleep(60 * time.Millisecond)
		exists, err := store.Exists(refreshKey)
		if err != nil {
			t.Fatalf("Exists for TTL refresh test returned error: %v", err)
		}
		if !exists {
			t.Fatal("expected HSetIfAbsent to refresh key TTL even when field already exists")
		}

		value, ok, err := store.HGet(key, "name")
		if err != nil {
			t.Fatalf("HGet returned error: %v", err)
		}
		if !ok || value != "boss" {
			t.Fatalf("expected HGet to return boss, got ok=%v value=%q", ok, value)
		}

		ok, err = store.HExists(key, "role")
		if err != nil {
			t.Fatalf("HExists returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected HExists to find role field")
		}

		length, err := store.HLen(key)
		if err != nil {
			t.Fatalf("HLen returned error: %v", err)
		}
		if length != 2 {
			t.Fatalf("expected HLen to return 2, got %d", length)
		}

		all, err := store.HGetAll(key)
		if err != nil {
			t.Fatalf("HGetAll returned error: %v", err)
		}
		if len(all) != 2 || all["name"] != "boss" || all["role"] != "admin" {
			t.Fatalf("unexpected HGetAll result: %+v", all)
		}

		total, err := store.HIncr(key, "count", 2, 0)
		if err != nil {
			t.Fatalf("first HIncr returned error: %v", err)
		}
		if total != 2 {
			t.Fatalf("expected first HIncr to return 2, got %d", total)
		}

		total, err = store.HIncr(key, "count", 3, 0)
		if err != nil {
			t.Fatalf("second HIncr returned error: %v", err)
		}
		if total != 5 {
			t.Fatalf("expected second HIncr to return 5, got %d", total)
		}

		if _, err := store.HSet(key, "broken", "abc", 0); err != nil {
			t.Fatalf("HSet for broken field returned error: %v", err)
		}
		if _, err := store.HIncr(key, "broken", 1, 0); err == nil {
			t.Fatal("expected HIncr on non-numeric field to fail")
		}

		deleted, err := store.HDel(key, "name", "role", "count", "broken")
		if err != nil {
			t.Fatalf("HDel returned error: %v", err)
		}
		if deleted != 4 {
			t.Fatalf("expected HDel to delete 4 fields, got %d", deleted)
		}

		exists, err = store.Exists(key)
		if err != nil {
			t.Fatalf("Exists after HDel returned error: %v", err)
		}
		if exists {
			t.Fatal("expected empty hash key to be deleted")
		}

		stringKey := testKey(factory.name, "hash-wrong-type")
		if err := store.Set(stringKey, "string", time.Second); err != nil {
			t.Fatalf("Set for hash wrong-type test returned error: %v", err)
		}
		if _, _, err := store.HGet(stringKey, "field"); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected HGet on string key to return ErrWrongType, got %v", err)
		}
	})

	t.Run(factory.name+"/set_ops", func(t *testing.T) {
		if _, err := store.SAdd(testKey(factory.name, "set-empty"), time.Second); err == nil {
			t.Fatal("expected SAdd without members to fail")
		}

		key := testKey(factory.name, "set")
		added, err := store.SAdd(key, 120*time.Millisecond, "a", "b", "b", "c")
		if err != nil {
			t.Fatalf("SAdd returned error: %v", err)
		}
		if added != 3 {
			t.Fatalf("expected SAdd to add 3 unique members, got %d", added)
		}

		ok, err := store.SContains(key, "b")
		if err != nil {
			t.Fatalf("SContains returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected SContains to find member b")
		}

		members, err := store.SMembers(key)
		if err != nil {
			t.Fatalf("SMembers returned error: %v", err)
		}
		sort.Strings(members)
		if len(members) != 3 || members[0] != "a" || members[2] != "c" {
			t.Fatalf("unexpected SMembers result: %+v", members)
		}

		card, err := store.SCard(key)
		if err != nil {
			t.Fatalf("SCard returned error: %v", err)
		}
		if card != 3 {
			t.Fatalf("expected SCard to return 3, got %d", card)
		}

		popped, err := store.SPop(key, 2)
		if err != nil {
			t.Fatalf("SPop returned error: %v", err)
		}
		if len(popped) != 2 || popped[0] == popped[1] {
			t.Fatalf("expected SPop to return 2 unique members, got %+v", popped)
		}

		card, err = store.SCard(key)
		if err != nil {
			t.Fatalf("SCard after SPop returned error: %v", err)
		}
		if card != 1 {
			t.Fatalf("expected SCard after SPop to return 1, got %d", card)
		}

		removed, err := store.SRem(key, "a", "b", "c")
		if err != nil {
			t.Fatalf("SRem returned error: %v", err)
		}
		if removed != 1 {
			t.Fatalf("expected SRem to remove the remaining member, got %d", removed)
		}

		exists, err := store.Exists(key)
		if err != nil {
			t.Fatalf("Exists after SRem returned error: %v", err)
		}
		if exists {
			t.Fatal("expected empty set key to be deleted")
		}

		stringKey := testKey(factory.name, "set-wrong-type")
		if err := store.Set(stringKey, "string", time.Second); err != nil {
			t.Fatalf("Set for set wrong-type test returned error: %v", err)
		}
		if _, err := store.SMembers(stringKey); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected SMembers on string key to return ErrWrongType, got %v", err)
		}
	})

	t.Run(factory.name+"/list_ops", func(t *testing.T) {
		if _, err := store.LPush(testKey(factory.name, "list-empty-left"), time.Second); err == nil {
			t.Fatal("expected LPush without values to fail")
		}
		if _, err := store.RPush(testKey(factory.name, "list-empty-right"), time.Second); err == nil {
			t.Fatal("expected RPush without values to fail")
		}

		key := testKey(factory.name, "list")
		length, err := store.LPush(key, 120*time.Millisecond, "a", "b")
		if err != nil {
			t.Fatalf("LPush returned error: %v", err)
		}
		if length != 2 {
			t.Fatalf("expected LPush length 2, got %d", length)
		}

		length, err = store.RPush(key, 120*time.Millisecond, "c", "d")
		if err != nil {
			t.Fatalf("RPush returned error: %v", err)
		}
		if length != 4 {
			t.Fatalf("expected RPush length 4, got %d", length)
		}

		values, err := store.LRange(key, 0, -1)
		if err != nil {
			t.Fatalf("LRange returned error: %v", err)
		}
		expected := []string{"b", "a", "c", "d"}
		if fmt.Sprint(values) != fmt.Sprint(expected) {
			t.Fatalf("expected %v, got %v", expected, values)
		}

		value, ok, err := store.LPop(key)
		if err != nil {
			t.Fatalf("LPop returned error: %v", err)
		}
		if !ok || value != "b" {
			t.Fatalf("expected LPop to return b, got ok=%v value=%q", ok, value)
		}

		value, ok, err = store.RPop(key)
		if err != nil {
			t.Fatalf("RPop returned error: %v", err)
		}
		if !ok || value != "d" {
			t.Fatalf("expected RPop to return d, got ok=%v value=%q", ok, value)
		}

		length, err = store.LLen(key)
		if err != nil {
			t.Fatalf("LLen returned error: %v", err)
		}
		if length != 2 {
			t.Fatalf("expected LLen to return 2, got %d", length)
		}

		if err := store.LTrim(key, 0, 0); err != nil {
			t.Fatalf("LTrim returned error: %v", err)
		}
		values, err = store.LRange(key, 0, -1)
		if err != nil {
			t.Fatalf("LRange after LTrim returned error: %v", err)
		}
		if fmt.Sprint(values) != fmt.Sprint([]string{"a"}) {
			t.Fatalf("expected trimmed list [a], got %v", values)
		}

		stringKey := testKey(factory.name, "list-wrong-type")
		if err := store.Set(stringKey, "string", time.Second); err != nil {
			t.Fatalf("Set for list wrong-type test returned error: %v", err)
		}
		if _, err := store.LRange(stringKey, 0, -1); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected LRange on string key to return ErrWrongType, got %v", err)
		}
	})

	t.Run(factory.name+"/sorted_set_ops", func(t *testing.T) {
		if _, err := store.ZAdd(testKey(factory.name, "zset-empty"), time.Second); err == nil {
			t.Fatal("expected ZAdd without members to fail")
		}

		key := testKey(factory.name, "zset")
		added, err := store.ZAdd(key, 120*time.Millisecond,
			ScoredMember{Member: "b", Score: 1},
			ScoredMember{Member: "a", Score: 1},
			ScoredMember{Member: "c", Score: 2},
		)
		if err != nil {
			t.Fatalf("ZAdd returned error: %v", err)
		}
		if added != 3 {
			t.Fatalf("expected ZAdd to add 3 members, got %d", added)
		}

		items, err := store.ZRange(key, 0, -1)
		if err != nil {
			t.Fatalf("ZRange returned error: %v", err)
		}
		expectMembers(t, items, []string{"a", "b", "c"})

		items, err = store.ZRevRange(key, 0, -1)
		if err != nil {
			t.Fatalf("ZRevRange returned error: %v", err)
		}
		expectMembers(t, items, []string{"c", "b", "a"})

		score, ok, err := store.ZScore(key, "a")
		if err != nil {
			t.Fatalf("ZScore returned error: %v", err)
		}
		if !ok || score != 1 {
			t.Fatalf("expected ZScore to return 1, got ok=%v score=%v", ok, score)
		}

		card, err := store.ZCard(key)
		if err != nil {
			t.Fatalf("ZCard returned error: %v", err)
		}
		if card != 3 {
			t.Fatalf("expected ZCard to return 3, got %d", card)
		}

		items, err = store.ZRangeByScore(key, 1, 1, 1)
		if err != nil {
			t.Fatalf("ZRangeByScore returned error: %v", err)
		}
		expectMembers(t, items, []string{"a"})

		items, err = store.ZPopMin(key, 2)
		if err != nil {
			t.Fatalf("ZPopMin returned error: %v", err)
		}
		expectMembers(t, items, []string{"a", "b"})

		items, err = store.ZPopMax(key, 1)
		if err != nil {
			t.Fatalf("ZPopMax returned error: %v", err)
		}
		expectMembers(t, items, []string{"c"})

		exists, err := store.Exists(key)
		if err != nil {
			t.Fatalf("Exists after ZPop returned error: %v", err)
		}
		if exists {
			t.Fatal("expected sorted set key to be deleted after all members were popped")
		}

		stringKey := testKey(factory.name, "zset-wrong-type")
		if err := store.Set(stringKey, "string", time.Second); err != nil {
			t.Fatalf("Set for zset wrong-type test returned error: %v", err)
		}
		if _, err := store.ZRange(stringKey, 0, -1); !errors.Is(err, ErrWrongType) {
			t.Fatalf("expected ZRange on string key to return ErrWrongType, got %v", err)
		}
	})

	t.Run(factory.name+"/lock_refresh", func(t *testing.T) {
		key := testKey(factory.name, "lock")
		owner := store.NewLock(key)
		if !owner.TryLock(100 * time.Millisecond) {
			t.Fatal("expected owner to acquire lock")
		}

		other := store.NewLock(key)
		ok, err := other.Refresh(150 * time.Millisecond)
		if err != nil {
			t.Fatalf("non-owner Refresh returned error: %v", err)
		}
		if ok {
			t.Fatal("expected non-owner Refresh to fail")
		}

		time.Sleep(60 * time.Millisecond)
		ok, err = owner.Refresh(150 * time.Millisecond)
		if err != nil {
			t.Fatalf("owner Refresh returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected owner Refresh to succeed")
		}

		time.Sleep(70 * time.Millisecond)
		challenger := store.NewLock(key)
		if challenger.TryLock(80 * time.Millisecond) {
			t.Fatal("expected refreshed lock to still be held")
		}

		time.Sleep(100 * time.Millisecond)
		if !challenger.TryLock(80 * time.Millisecond) {
			t.Fatal("expected refreshed lock to expire eventually")
		}
	})
}

func TestLocalKVSetIfAbsentConcurrent(t *testing.T) {
	store := newLocalKV()
	key := testKey("local", "race")

	var successCount atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ok, err := store.SetIfAbsent(key, "winner", 80*time.Millisecond)
			if err != nil {
				t.Errorf("SetIfAbsent returned error: %v", err)
				return
			}
			if ok {
				successCount.Add(1)
			}
		}()
	}

	close(start)
	wg.Wait()

	if successCount.Load() != 1 {
		t.Fatalf("expected exactly one SetIfAbsent winner, got %d", successCount.Load())
	}
}

func expectMembers(t *testing.T, items []ScoredMember, expected []string) {
	t.Helper()
	members := make([]string, 0, len(items))
	for _, item := range items {
		members = append(members, item.Member)
	}
	if fmt.Sprint(members) != fmt.Sprint(expected) {
		t.Fatalf("expected members %v, got %v", expected, members)
	}
}

func testKey(scope string, name string) string {
	return fmt.Sprintf("allingo:%s:%s:%d", scope, name, time.Now().UnixNano())
}
