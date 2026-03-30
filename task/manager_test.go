package task

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/real-uangi/allingo/kv"
)

func TestTaskOccupyUsesAtomicKVOps(t *testing.T) {
	store := newTaskTestKV()

	first := &Task{
		taskId:  "COCK_CLIENT:TASK:test-occupy",
		counter: 0,
		kv:      store,
	}
	second := &Task{
		taskId:  first.taskId,
		counter: 0,
		kv:      store,
	}

	var winners atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})

	runRound := func(task *Task) {
		defer wg.Done()
		<-start
		if task.occupy() {
			winners.Add(1)
		}
	}

	wg.Add(2)
	go runRound(first)
	go runRound(second)
	close(start)
	wg.Wait()

	if got := winners.Load(); got != 1 {
		t.Fatalf("expected one winner in first round, got %d", got)
	}
	if first.counter != 1 || second.counter != 1 {
		t.Fatalf("expected both tasks to sync counter to 1, got first=%d second=%d", first.counter, second.counter)
	}

	winners.Store(0)
	start = make(chan struct{})

	wg.Add(2)
	go runRound(first)
	go runRound(second)
	close(start)
	wg.Wait()

	if got := winners.Load(); got != 1 {
		t.Fatalf("expected one winner in second round, got %d", got)
	}
	if first.counter != 2 || second.counter != 2 {
		t.Fatalf("expected both tasks to sync counter to 2, got first=%d second=%d", first.counter, second.counter)
	}
}

type taskTestKV struct {
	mu   sync.Mutex
	data map[string]string
}

func newTaskTestKV() *taskTestKV {
	return &taskTestKV{
		data: make(map[string]string),
	}
}

func (kv *taskTestKV) Exists(key string) (bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	_, ok := kv.data[key]
	return ok, nil
}

func (kv *taskTestKV) Del(key string) (bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	_, ok := kv.data[key]
	delete(kv.data, key)
	return ok, nil
}

func (kv *taskTestKV) Expire(key string, ttl time.Duration) (bool, error) {
	return false, nil
}

func (kv *taskTestKV) Persist(key string) (bool, error) {
	return false, nil
}

func (kv *taskTestKV) TTL(key string) (time.Duration, bool, error) {
	return 0, false, nil
}

func (kv *taskTestKV) Set(key string, value string, ttl time.Duration) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[key] = value
	return nil
}

func (kv *taskTestKV) Get(key string) (string, bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	value, ok := kv.data[key]
	return value, ok, nil
}

func (kv *taskTestKV) SetIfAbsent(key string, value string, ttl time.Duration) (bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if _, ok := kv.data[key]; ok {
		return false, nil
	}
	kv.data[key] = value
	return true, nil
}

func (kv *taskTestKV) SetIfPresent(key string, value string, ttl time.Duration) (bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if _, ok := kv.data[key]; !ok {
		return false, nil
	}
	kv.data[key] = value
	return true, nil
}

func (kv *taskTestKV) CompareAndSet(key string, expected, value string, ttl time.Duration) (bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.data[key] != expected {
		return false, nil
	}
	kv.data[key] = value
	return true, nil
}

func (kv *taskTestKV) CompareAndDelete(key string, expected string) (bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.data[key] != expected {
		return false, nil
	}
	delete(kv.data, key)
	return true, nil
}

func (kv *taskTestKV) GetAndDelete(key string) (string, bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	value, ok := kv.data[key]
	if ok {
		delete(kv.data, key)
	}
	return value, ok, nil
}

func (kv *taskTestKV) GetAndSet(key string, value string, ttl time.Duration) (string, bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	old, ok := kv.data[key]
	kv.data[key] = value
	return old, ok, nil
}

func (kv *taskTestKV) MGet(keys ...string) (map[string]string, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := kv.data[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (kv *taskTestKV) MSet(values map[string]string, ttl time.Duration) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	for key, value := range values {
		kv.data[key] = value
	}
	return nil
}

func (kv *taskTestKV) MDel(keys ...string) (int64, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	var deleted int64
	for _, key := range keys {
		if _, ok := kv.data[key]; ok {
			delete(kv.data, key)
			deleted++
		}
	}
	return deleted, nil
}

func (kv *taskTestKV) SetStruct(key string, obj any, ttl time.Duration) error {
	return nil
}

func (kv *taskTestKV) GetStruct(key string, p any) (bool, error) {
	return false, nil
}

func (kv *taskTestKV) HSet(key string, field string, value string, ttl time.Duration) (bool, error) {
	return false, nil
}

func (kv *taskTestKV) HSetIfAbsent(key string, field string, value string, ttl time.Duration) (bool, error) {
	return false, nil
}

func (kv *taskTestKV) HGet(key string, field string) (string, bool, error) {
	return "", false, nil
}

func (kv *taskTestKV) HDel(key string, fields ...string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) HExists(key string, field string) (bool, error) {
	return false, nil
}

func (kv *taskTestKV) HGetAll(key string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (kv *taskTestKV) HLen(key string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) HIncr(key string, field string, delta int64, ttl time.Duration) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) SAdd(key string, ttl time.Duration, members ...string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) SRem(key string, members ...string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) SContains(key string, member string) (bool, error) {
	return false, nil
}

func (kv *taskTestKV) SMembers(key string) ([]string, error) {
	return []string{}, nil
}

func (kv *taskTestKV) SCard(key string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) SPop(key string, count int64) ([]string, error) {
	return []string{}, nil
}

func (kv *taskTestKV) LPush(key string, ttl time.Duration, values ...string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) RPush(key string, ttl time.Duration, values ...string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) LPop(key string) (string, bool, error) {
	return "", false, nil
}

func (kv *taskTestKV) RPop(key string) (string, bool, error) {
	return "", false, nil
}

func (kv *taskTestKV) LLen(key string) (int64, error) {
	return 0, nil
}

func (kv *taskTestKV) LRange(key string, start, stop int64) ([]string, error) {
	return []string{}, nil
}

func (kv *taskTestKV) LTrim(key string, start, stop int64) error {
	return nil
}

func (store *taskTestKV) ZAdd(key string, ttl time.Duration, members ...kv.ScoredMember) (int64, error) {
	return 0, nil
}

func (store *taskTestKV) ZRem(key string, members ...string) (int64, error) {
	return 0, nil
}

func (store *taskTestKV) ZScore(key string, member string) (float64, bool, error) {
	return 0, false, nil
}

func (store *taskTestKV) ZCard(key string) (int64, error) {
	return 0, nil
}

func (store *taskTestKV) ZRange(key string, start, stop int64) ([]kv.ScoredMember, error) {
	return []kv.ScoredMember{}, nil
}

func (store *taskTestKV) ZRevRange(key string, start, stop int64) ([]kv.ScoredMember, error) {
	return []kv.ScoredMember{}, nil
}

func (store *taskTestKV) ZRangeByScore(key string, min, max float64, limit int64) ([]kv.ScoredMember, error) {
	return []kv.ScoredMember{}, nil
}

func (store *taskTestKV) ZPopMin(key string, count int64) ([]kv.ScoredMember, error) {
	return []kv.ScoredMember{}, nil
}

func (store *taskTestKV) ZPopMax(key string, count int64) ([]kv.ScoredMember, error) {
	return []kv.ScoredMember{}, nil
}

func (store *taskTestKV) GetType() kv.Type {
	return kv.Local
}

func (kv *taskTestKV) NewLock(key string) kv.Lock {
	panic("task manager should not require NewLock for occupy")
}

func (kv *taskTestKV) Ping() error {
	return nil
}

func (kv *taskTestKV) Incr(key string, i int64, ttl time.Duration, nx bool) (int64, error) {
	return 0, nil
}
