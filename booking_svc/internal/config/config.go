package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName     string
	HTTPPort        string
	GracefulTimeout time.Duration
	LogLevel        string
}

func LoadFromEnv(serviceName, defaultPort string) Config {
	logLevel := getEnv("LOG_LEVEL", "info")
	port := getEnv("HTTP_PORT", defaultPort)
	gt := getEnvInt("GRACEFUL_TIMEOUT_SECONDS", 10)

	return Config{
		ServiceName:     serviceName,
		HTTPPort:        port,
		GracefulTimeout: time.Duration(gt) * time.Second,
		LogLevel:        logLevel,
	}
}

func (c Config) Addr() string {
	return ":" + c.HTTPPort
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
