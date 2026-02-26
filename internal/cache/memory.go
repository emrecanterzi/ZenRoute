package cache

import "sync"

type InMemoryCache struct {
	cache map[string]string
	mu    sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]string),
	}
}

func (c *InMemoryCache) Set(domain string, ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[domain] = ip
}

func (c *InMemoryCache) Get(domain string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if ip, ok := c.cache[domain]; ok {
		return ip, true
	}
	return "", false
}
