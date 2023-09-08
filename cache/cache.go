// Package cache contains caching implementation for retrieved data
package cache

import (
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// ErrKeyNotFound signalized the key is not present in the cache
var ErrKeyNotFound = errors.New("requested key was not found in database")

// Cache describes an interface used to store generated data
type Cache interface {
	Get(namespace, key string) (value string, err error)
	Set(namespace, key, value string, ttl time.Duration) (err error)
	Delete(namespace, key string) (err error)
}

// GetCacheByURI instantiates a new Cache by the given URI string
func GetCacheByURI(uri string) (Cache, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, errors.Wrap(err, "parsing uri")
	}

	switch u.Scheme {
	case "mem":
		return NewInMemCache(), nil
	default:
		return nil, errors.New("Invalid cache scheme: " + u.Scheme)
	}
}
