package cache

type Cache interface {
	Set(domain string, ip string)
	Get(domain string) (string, bool)
}
