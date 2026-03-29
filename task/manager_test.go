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

func (kv *taskTestKV) Set(key string, value string, ttl time.Duration) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[key] = value
}

func (kv *taskTestKV) Get(key string) (string, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	value, ok := kv.data[key]
	return value, ok
}

func (kv *taskTestKV) Del(key string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.data, key)
	return nil
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

func (kv *taskTestKV) SetStruct(key string, obj interface{}, ttl time.Duration) error {
	return nil
}

func (kv *taskTestKV) GetStruct(key string, p any) error {
	return nil
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
