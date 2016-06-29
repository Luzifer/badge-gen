package cache

import (
	"errors"
	"net/url"
	"time"
)

type KeyNotFoundError struct{}

func (k KeyNotFoundError) Error() string {
	return "Requested key was not found in database"
}

type Cache interface {
	Get(namespace, key string) (value string, err error)
	Set(namespace, key, value string, ttl time.Duration) (err error)
	Delete(namespace, key string) (err error)
}

func GetCacheByURI(uri string) (Cache, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "mem":
		return NewInMemCache(), nil
	default:
		return nil, errors.New("Invalid cache scheme: " + url.Scheme)
	}
}
