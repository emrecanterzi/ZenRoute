package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ProxyPort         string
	SystemServiceName string
	FragmentSize      int
	BypassDomains     []string
	BypassAll         bool
	LocalOnly         bool
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
		return nil, fmt.Errorf("bypass domains: %w", err)
	}

	localOnly := getEnv("LOCAL_ONLY", "false") == "true"

	return &Config{
		ProxyPort:         getEnv("PROXY_PORT", "8080"),
		SystemServiceName: getEnv("SYSTEM_SERVICE", "Wi-Fi"),
		FragmentSize:      getEnvInt("FRAGMENT_SIZE", 7),
		BypassDomains:     domains,
		BypassAll:         getEnv("BYPASS_ALL", "false") == "true",
		LocalOnly:         localOnly,
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
