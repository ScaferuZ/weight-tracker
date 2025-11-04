package config

import (
	"os"
)

type Config struct {
	Port       string
	DatabasePath string
	Env       string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabasePath: getEnv("DB_PATH", "./data/weights.db"),
		Env:        getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}