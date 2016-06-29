package cache

import (
	"sync"
	"time"
)

type inMemCacheEntry struct {
	Value   string
	Expires time.Time
}

type InMemCache struct {
	cache map[string]inMemCacheEntry
	lock  sync.RWMutex
}

func NewInMemCache() *InMemCache {
	return &InMemCache{
		cache: map[string]inMemCacheEntry{},
	}
}

func (i InMemCache) Get(namespace, key string) (value string, err error) {
	i.lock.RLock()
	defer i.lock.RUnlock()

	e, ok := i.cache[namespace+"::"+key]
	if !ok || e.Expires.Before(time.Now()) {
		return "", KeyNotFoundError{}
	}
	return e.Value, nil
}

func (i InMemCache) Set(namespace, key, value string, ttl time.Duration) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.cache[namespace+"::"+key] = inMemCacheEntry{
		Value:   value,
		Expires: time.Now().Add(ttl),
	}

	return nil
}

func (i InMemCache) Delete(namespace, key string) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.cache, namespace+"::"+key)
	return nil
}
