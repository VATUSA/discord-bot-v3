package config

import "os"

var HTTPAddr = getEnvOrDefault("HTTP_ADDR", ":9002")

func getEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
