package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	RedisAddr     string
	RedisPassword string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	ServerPort int
	ServerHost string

	LogLevel string

	YahooFinanceEnabled bool
	PriceFetchInterval  time.Duration
	LookupCacheTTL      time.Duration
	YahooMinInterval    time.Duration
	YahooGlobalRate     int
	YahooGlobalWindow   time.Duration
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("VAULT_DB_HOST", "localhost"),
		DBPort:     getEnvInt("VAULT_DB_PORT", 5432),
		DBName:     getEnv("VAULT_DB_NAME", "vaultlab"),
		DBUser:     getEnv("VAULT_DB_USER", "vaultlab"),
		DBPassword: getEnv("VAULT_DB_PASSWORD", "vaultlab"),
		DBSSLMode:  getEnv("VAULT_DB_SSLMODE", "disable"),

		RedisAddr:     getEnv("VAULT_REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("VAULT_REDIS_PASSWORD", ""),

		JWTSecret:     getEnv("VAULT_JWT_SECRET", "change-me-in-production"),
		JWTAccessTTL:  getEnvDuration("VAULT_JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getEnvDuration("VAULT_JWT_REFRESH_TTL", 72*time.Hour),

		ServerPort: getEnvInt("VAULT_SERVER_PORT", 8080),
		ServerHost: getEnv("VAULT_SERVER_HOST", "0.0.0.0"),

		LogLevel: getEnv("VAULT_LOG_LEVEL", "debug"),

		YahooFinanceEnabled: getEnvBool("VAULT_YAHOO_FINANCE_ENABLED", true),
		PriceFetchInterval:  getEnvDuration("VAULT_PRICE_FETCH_INTERVAL", 1*time.Hour),
		LookupCacheTTL:      getEnvDuration("VAULT_LOOKUP_CACHE_TTL", 7*24*time.Hour),
		YahooMinInterval:    getEnvDuration("VAULT_YAHOO_MIN_INTERVAL", 400*time.Millisecond),
		YahooGlobalRate:     getEnvInt("VAULT_YAHOO_GLOBAL_RATE", 8),
		YahooGlobalWindow:   getEnvDuration("VAULT_YAHOO_GLOBAL_WINDOW", 1*time.Second),
	}
}

func (c *Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + strconv.Itoa(c.DBPort) + "/" + c.DBName + "?sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
