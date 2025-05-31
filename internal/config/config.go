package config

import "os"

type Config struct {
    Port         string
    DatabasePath string
}

func Load() *Config {
    return &Config{
        Port:         getEnv("PORT", "8080"),
        DatabasePath: getEnv("DATABASE_PATH", "stampkeeper.db"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}