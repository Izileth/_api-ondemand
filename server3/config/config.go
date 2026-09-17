package config

import "os"

const (
	DefaultPort       = "8082"
	DefaultServerName = "Server 3"
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
