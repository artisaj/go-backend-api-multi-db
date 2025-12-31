package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config centraliza parâmetros de execução carregados via variáveis de ambiente.
type Config struct {
	Env        string
	Port       int
	LogLevel   string
	Mongo      MongoConfig
	Cache      CacheConfig
	Thresholds ThresholdsConfig
}

// MongoConfig define onde ficam os metadados de fontes de dados.
type MongoConfig struct {
	URI    string
	DBName string
}

// CacheConfig controla TTL e limites de cache.
type CacheConfig struct {
	TTLSeconds int
	MaxItems   int
}

// ThresholdsConfig determina limites de tempo e custo para consultas.
type ThresholdsConfig struct {
	QueryTimeoutMs   int
	AsyncSwitchP95Ms int
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Env:        getEnv("NODE_ENV", "development"),
		Port:       intFromEnv("PORT", 8080),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		Mongo:      loadMongo(),
		Cache:      loadCache(),
		Thresholds: loadThresholds(),
	}
}

func loadMongo() MongoConfig {
	return MongoConfig{
		URI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName: getEnv("MONGO_DB_NAME", "api_database_config"),
	}
}

func loadCache() CacheConfig {
	return CacheConfig{
		TTLSeconds: intFromEnv("CACHE_TTL_SECONDS", 300),
		MaxItems:   intFromEnv("CACHE_MAX_ITEMS", 1000),
	}
}

func loadThresholds() ThresholdsConfig {
	return ThresholdsConfig{
		QueryTimeoutMs:   intFromEnv("QUERY_TIMEOUT_MS", 4000),
		AsyncSwitchP95Ms: intFromEnv("ASYNC_SWITCH_P95_MS", 3500),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intFromEnv(key string, fallback int) int {
	value := os.Getenv(key)
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
