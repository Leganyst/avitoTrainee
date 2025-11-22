package config

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	LogLevel string
}

var appLogger *zap.SugaredLogger

func Load() *Config {
	return &Config{
		Port:     getEnv("APP_PORT", "8080"),
		DBHost:   getEnv("DB_HOST", "localhost"),
		DBPort:   getEnv("DB_PORT", "5432"),
		DBUser:   getEnv("DB_USER", "app"),
		DBPass:   getEnv("DB_PASS", "app"),
		DBName:   getEnv("DB_NAME", "app"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func InitLogger(level string) error {
	var cfg zap.Config
	switch strings.ToLower(level) {
	case "debug":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewProductionConfig()
	}
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if strings.ToLower(level) == "debug" {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	appLogger = logger.Sugar()
	return nil
}

func Logger() *zap.SugaredLogger {
	return appLogger
}
