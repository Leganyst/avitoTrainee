package config

import (
	"os"
)

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func Load() *Config {
	c := &Config{
		Port:   getEnv("APP_PORT", "8080"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "app"),
		DBPass: getEnv("DB_PASS", "app"),
		DBName: getEnv("DB_NAME", "app"),
	}

	return c
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
