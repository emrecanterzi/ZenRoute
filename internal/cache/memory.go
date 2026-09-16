package cache

import (
	"sync"
	"time"
)

type entry struct {
	ip        string
	expiresAt time.Time
}

type InMemoryCache struct {
	cache map[string]entry
	mu    sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]entry),
	}
}

func (c *InMemoryCache) Set(domain, ip string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[domain] = entry{ip: ip, expiresAt: time.Now().Add(ttl)}
}

func (c *InMemoryCache) Get(domain string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.cache[domain]
	if !ok {
		return "", false
	}
	if time.Now().After(e.expiresAt) {
		delete(c.cache, domain)
		return "", false
	}
	return e.ip, true
}
