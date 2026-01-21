package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr   string
	ValkeyAddrs []string
	ValkeyUser string
	ValkeyPass string
	IndexName  string
	KeyPrefix  string
}

func FromEnv() Config {
	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		ValkeyAddrs: splitCSV(getenv("VALKEY_ADDRS", "localhost:6379")),
		ValkeyUser:  os.Getenv("VALKEY_USER"),
		ValkeyPass:  os.Getenv("VALKEY_PASS"),
		IndexName:   getenv("VALKEY_INDEX", "index_places"),
		KeyPrefix:   getenv("VALKEY_PREFIX", "places:"),
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		out = []string{"localhost:6379"}
	}
	return out
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
