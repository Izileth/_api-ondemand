package config

import "os"

const (
	DefaultPort       = "8080"
	DefaultServerName = "Server 1"
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
