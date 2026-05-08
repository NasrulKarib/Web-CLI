package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	Host          string
	ShellDefault  string
	ShellTimeout  int
	LogBufferSize int
}

func Load() *Config {
	// Load .env file if it exists (optional, won't fail if missing)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v (using system environment variables)", err)
	}

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
