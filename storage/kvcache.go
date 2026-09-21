package storage

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type KVCache[V any] interface {
	Set(ctx context.Context, key string, value V, ttl time.Duration) error
	Get(ctx context.Context, key string) (value V, found bool, err error)
	Delete(ctx context.Context, key string) (found bool, err error)
	Close() error
}

const MemoryJanitorInterval = time.Minute

type MemoryEntry[V any] struct {
	value   V
	expires time.Time
}

func (e MemoryEntry[V]) Expired(now time.Time) bool {
	return !e.expires.IsZero() && now.After(e.expires)
}

type MemoryKVCache[V any] struct {
	entries   sync.Map
	done      chan struct{}
	closeOnce sync.Once
}

func NewMemoryKVCache[V any]() *MemoryKVCache[V] {
	m := &MemoryKVCache[V]{done: make(chan struct{})}
	go m.janitor()
	return m
}

func (m *MemoryKVCache[V]) Set(_ context.Context, key string, value V, ttl time.Duration) error {
	m.entries.Store(key, MemoryEntry[V]{value: value, expires: time.Now().Add(ttl)})
	return nil
}

func (m *MemoryKVCache[V]) Get(_ context.Context, key string) (V, bool, error) {
	raw, ok := m.entries.Load(key)
	if !ok {
		var zero V
		return zero, false, nil
	}
	entry, ok := raw.(MemoryEntry[V])
	if !ok {
		var zero V
		return zero, false, nil
	}
	if entry.Expired(time.Now()) {
		m.entries.CompareAndDelete(key, raw)
		var zero V
		return zero, false, nil
	}
	return entry.value, true, nil
}

func (m *MemoryKVCache[V]) Delete(_ context.Context, key string) (bool, error) {
	raw, ok := m.entries.LoadAndDelete(key)
	if !ok {
		return false, nil
	}
	entry, ok := raw.(MemoryEntry[V])
	if !ok {
		return false, nil
	}
	return !entry.Expired(time.Now()), nil
}

func (m *MemoryKVCache[V]) Close() error {
	m.closeOnce.Do(func() { close(m.done) })
	return nil
}

func (m *MemoryKVCache[V]) janitor() {
	ticker := time.NewTicker(MemoryJanitorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			m.entries.Range(func(key, raw any) bool {
				if entry, ok := raw.(MemoryEntry[V]); ok && entry.Expired(now) {
					m.entries.CompareAndDelete(key, raw)
				}
				return true
			})
		case <-m.done:
			return
		}
	}
}

type RedisKVCache struct {
	client *redis.Client
}

func NewRedisKVCache(ctx context.Context, addr, password string) (*RedisKVCache, error) {
	client := redis.NewClient(&redis.Options{Addr: addr, Password: password})
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return &RedisKVCache{client: client}, nil
}

func (r *RedisKVCache) Close() error {
	return r.client.Close()
}

func (r *RedisKVCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl < 0 {
		ttl = 0
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisKVCache) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (r *RedisKVCache) Delete(ctx context.Context, key string) (bool, error) {
	_, err := r.client.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
