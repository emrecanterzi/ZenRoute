package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ProxyAddr         string
	ProxyPort         string
	SystemServiceName string
	FragmentSize      int
	BypassDomains     []string
	BypassAll         bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	domainsFile := getEnv("BYPASS_DOMAINS_FILE", "./bypass-domains.txt")
	domains, err := loadBypassDomains(domainsFile)
	if err != nil {
		return nil, fmt.Errorf("bypass domains: %w", err)
	}

	return &Config{
		ProxyAddr:         getEnv("PROXY_ADDR", "127.0.0.1"),
		ProxyPort:         getEnv("PROXY_PORT", "8080"),
		SystemServiceName: getEnv("SYSTEM_SERVICE", "Wi-Fi"),
		FragmentSize:      getEnvInt("FRAGMENT_SIZE", 7),
		BypassDomains:     domains,
		BypassAll:         getEnv("BYPASS_ALL", "false") == "true",
	}, nil
}

func loadBypassDomains(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var domains []string
	for _, line := range strings.Split(string(data), "\n") {
		d := strings.TrimSpace(line)
		if d == "" || strings.HasPrefix(d, "#") {
			continue
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
