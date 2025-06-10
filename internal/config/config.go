package config

import (
    "fmt"
    "os"
)

type Config struct {
    Port           string
    DatabaseURL    string
}

func Load() *Config {
    // Build PostgreSQL connection string from environment variables
    host := getEnv("DB_HOST", "postgres") // Default to Docker service name
    port := getEnv("DB_PORT", "5432")
    user := getEnv("DB_USER", "bird")
    password := getEnv("DB_PASSWORD", "birder13")
    dbname := getEnv("DB_NAME", "stampkeeper")
    sslmode := getEnv("DB_SSLMODE", "disable")
    
    dbURL := getEnv("DATABASE_URL", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", 
        host, port, user, password, dbname, sslmode))
    
    return &Config{
        Port:        getEnv("PORT", "8080"),
        DatabaseURL: dbURL,
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}