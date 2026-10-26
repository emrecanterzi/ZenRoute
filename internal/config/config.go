package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emrecanterzi/zenroute/internal/util"
	"github.com/joho/godotenv"
)

type Config struct {
	ProxyPort         string
	SystemServiceName string
	BypassDomains     []string
	BypassAll         bool
	LocalOnly         bool
	MaxConns          int
}

func Load() (*Config, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	envFile := filepath.Join(execDir, ".env")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		envFile = ".env"
	}
	_ = godotenv.Load(envFile)

	domainsFile := filepath.Join(execDir, "bypass-domains.txt")
	if _, err := os.Stat(domainsFile); os.IsNotExist(err) {
		domainsFile = "./bypass-domains.txt"
	}
	domainsFile = getEnv("BYPASS_DOMAINS_FILE", domainsFile)

	domains, err := loadBypassDomains(domainsFile)
	if err != nil {
		fmt.Printf("zenroute: failed to load bypass domains: %v\n", err)
	}

	localOnly := getEnv("LOCAL_ONLY", "true") == "true"

	return &Config{
		ProxyPort:         getEnv("PROXY_PORT", "8080"),
		SystemServiceName: getEnv("SYSTEM_SERVICE", "Wi-Fi"),
		BypassDomains:     domains,
		BypassAll:         getEnv("BYPASS_ALL", "false") == "true" || len(domains) == 0,
		LocalOnly:         localOnly,
		MaxConns:          getEnvInt("MAX_CONNS", 1024, 1, 65535),
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
		domains = append(domains, util.NormalizeDomain(d))
	}
	return domains, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback, min, max int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			if i > max {
				fmt.Printf("zenroute: %s=%d above maximum, clamped to %d\n", key, i, max)
				return max
			}
			if i < min {
				fmt.Printf("zenroute: %s=%d below minimum, clamped to %d\n", key, i, min)
				return min
			}
			return i
		}
	}
	return fallback
}
