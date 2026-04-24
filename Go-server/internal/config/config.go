package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	Host          string
	ShellDefault  string
	ShellTimeout  int
	LogBufferSize int
}

func Load() *Config {
	return &Config{
		Port:          GetEnv("PORT", "8080"),
		Host:          GetEnv("HOST", "0.0.0.0"),
		ShellDefault:  GetEnv("SHELL_DEFAULT", "/bin/bash"),
		ShellTimeout:  GetEnvInt("SHELL_TIMEOUT", 3600),
		LogBufferSize: GetEnvInt("LOG_BUFFER_SIZE", 1000),
	}
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
