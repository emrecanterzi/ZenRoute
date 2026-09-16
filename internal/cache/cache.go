package cache

import "time"

type Cache interface {
	Set(domain string, ip string, ttl time.Duration)
	Get(domain string) (string, bool)
}
