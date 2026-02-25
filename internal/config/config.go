package config

import (
	"os"
)

type Config struct {
	ProxyAddr         string
	ProxyPort         string
	SystemServiceName string
	FragmentSize      int
}

func Load() *Config {
	return &Config{
		ProxyAddr:         getEnv("PROXY_ADDR", "127.0.0.1"),
		ProxyPort:         getEnv("PROXY_PORT", "8080"),
		SystemServiceName: getEnv("SYSTEM_SERVICE", "Wi-Fi"),
		FragmentSize:      7,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
