package config

import "os"

const (
	DefaultPort       = "8081"
	DefaultServerName = "Server 2"
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
