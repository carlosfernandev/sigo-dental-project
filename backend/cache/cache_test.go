package cache

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"path"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type fakeRedis struct {
	values  map[string][]byte
	ttls    map[string]time.Duration
	getErr  error
	setErr  error
	delErr  error
	scanErr error
}

func newFakeRedis() *fakeRedis {
	return &fakeRedis{values: make(map[string][]byte), ttls: make(map[string]time.Duration)}
}

func (r *fakeRedis) Get(_ context.Context, key string) ([]byte, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	value, ok := r.values[key]
	if !ok {
		return nil, redis.Nil
	}
	return value, nil
}

func (r *fakeRedis) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if r.setErr != nil {
		return r.setErr
	}
	r.values[key] = append([]byte(nil), value...)
	r.ttls[key] = ttl
	return nil
}

func (r *fakeRedis) Delete(_ context.Context, keys ...string) error {
	if r.delErr != nil {
		return r.delErr
	}
	for _, key := range keys {
		delete(r.values, key)
		delete(r.ttls, key)
	}
	return nil
}

func (r *fakeRedis) Scan(_ context.Context, _ uint64, pattern string, _ int64) ([]string, uint64, error) {
	if r.scanErr != nil {
		return nil, 0, r.scanErr
	}
	var keys []string
	for key := range r.values {
		matched, err := path.Match(pattern, key)
		if err == nil && matched {
			keys = append(keys, key)
		}
	}
	return keys, 0, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCacheRoundTripAndMiss(t *testing.T) {
	client := newFakeRedis()
	cache := &RedisCache{client: client, logger: testLogger()}
	ctx := context.Background()

	var got map[string]string
	if cache.Get(ctx, "patient:1", &got) {
		t.Fatal("Get() returned hit for missing key")
	}
	want := map[string]string{"status": "cached"}
	cache.Set(ctx, "patient:1", want, time.Minute)
	if !cache.Get(ctx, "patient:1", &got) {
		t.Fatal("Get() returned miss after Set()")
	}
	if got["status"] != want["status"] {
		t.Errorf("cached value = %#v, want %#v", got, want)
	}
	if client.ttls["patient:1"] != time.Minute {
		t.Errorf("TTL = %s, want 1m", client.ttls["patient:1"])
	}
}

func TestCacheTreatsRedisAndJSONErrorsAsMisses(t *testing.T) {
	client := newFakeRedis()
	cache := &RedisCache{client: client, logger: testLogger()}
	ctx := context.Background()

	client.getErr = errors.New("redis unavailable")
	if cache.Get(ctx, "any:key", new(any)) {
		t.Fatal("Get() returned hit while Redis is unavailable")
	}
	client.getErr = nil
	client.values["invalid:key"] = []byte("{")
	if cache.Get(ctx, "invalid:key", new(map[string]string)) {
		t.Fatal("Get() returned hit for invalid JSON")
	}
}

func TestCacheInvalidationAndDisabledClient(t *testing.T) {
	ctx := context.Background()
	client := newFakeRedis()
	cache := &RedisCache{client: client, logger: testLogger()}
	for _, key := range []string{"chair:1", "chair:2", "patient:1"} {
		client.values[key], _ = json.Marshal(key)
	}
	cache.Delete(ctx, "chair:1")
	cache.DeleteByPattern(ctx, "chair:*")
	if len(client.values) != 1 || client.values["patient:1"] == nil {
		t.Errorf("remaining keys = %v, want only patient:1", client.values)
	}

	disabled := NewRedisCache(nil, testLogger())
	if disabled.Get(ctx, "missing", new(string)) {
		t.Fatal("disabled cache returned a hit")
	}
	disabled.Set(ctx, "key", "value", time.Minute)
	disabled.Delete(ctx, "key")
	disabled.DeleteByPattern(ctx, "*")
}

func TestCacheRequiresPositiveTTL(t *testing.T) {
	client := newFakeRedis()
	cache := &RedisCache{client: client, logger: testLogger()}
	cache.Set(context.Background(), "no-ttl", "value", 0)
	if _, ok := client.values["no-ttl"]; ok {
		t.Fatal("Set() stored a value with no TTL")
	}
}
