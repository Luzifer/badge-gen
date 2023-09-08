package cache

import (
	"sync"
	"time"
)

type inMemCacheEntry struct {
	Value   string
	Expires time.Time
}

// InMemCache implements the Cache interface for storage in memory
type InMemCache struct {
	cache map[string]inMemCacheEntry
	lock  sync.RWMutex
}

// NewInMemCache creates a new InMemCache
func NewInMemCache() *InMemCache {
	return &InMemCache{
		cache: map[string]inMemCacheEntry{},
	}
}

// Get retrieves stored data
func (i *InMemCache) Get(namespace, key string) (value string, err error) {
	i.lock.RLock()
	defer i.lock.RUnlock()

	e, ok := i.cache[namespace+"::"+key]
	if !ok || e.Expires.Before(time.Now()) {
		return "", ErrKeyNotFound
	}
	return e.Value, nil
}

// Set stores data
func (i *InMemCache) Set(namespace, key, value string, ttl time.Duration) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.cache[namespace+"::"+key] = inMemCacheEntry{
		Value:   value,
		Expires: time.Now().Add(ttl),
	}

	return nil
}

// Delete deletes data
func (i *InMemCache) Delete(namespace, key string) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.cache, namespace+"::"+key)
	return nil
}
