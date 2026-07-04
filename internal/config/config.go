package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAppName           = "EVE Online DScan Tool"
	defaultAppVersion        = "2.0.0-dev"
	defaultHTTPAddr          = ":8000"
	defaultDatabaseURL       = "postgres://dscan:dscan@localhost:5432/dscan?sslmode=disable"
	defaultRedisAddr         = "localhost:6379"
	defaultESIBaseURL        = "https://esi.evetech.net/latest"
	defaultESIDatasource     = "tranquility"
	defaultESILanguage       = "en"
	defaultSDEURL            = ""
	defaultShortLinkLength   = 10
	defaultESIMaxRetries     = 3
	defaultESIRetryDelay     = time.Second
	defaultESIIDsBatchSize   = 500
	defaultESIBatchSize      = 1000
	defaultDScanCacheTTL     = time.Hour
	defaultESICharacterTTL   = 7 * 24 * time.Hour
	defaultESIAffiliationTTL = 7 * 24 * time.Hour
	defaultESIEntityNameTTL  = 30 * 24 * time.Hour
	defaultHTTPTimeout       = 30 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultDatabaseMaxConns  = int32(10)
	defaultDatabaseMinConns  = int32(1)
)

type Config struct {
	AppName    string
	AppVersion string
	AppEnv     string

	HTTPAddr        string
	HTTPTimeout     time.Duration
	ShutdownTimeout time.Duration

	DatabaseURL      string
	DatabaseMaxConns int32
	DatabaseMinConns int32

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	DScanCacheTTL     time.Duration
	ESICharacterTTL   time.Duration
	ESIAffiliationTTL time.Duration
	ESIEntityNameTTL  time.Duration

	ESIBaseURL      string
	ESIDatasource   string
	ESILanguage     string
	ESIMaxRetries   int
	ESIRetryDelay   time.Duration
	ESIIDsBatchSize int
	ESIBatchSize    int
	SDEURL          string

	ShortLinkLength int
}

func Load() Config {
	loadDotenvFiles()

	return Config{
		AppName:    getEnv("APP_NAME", defaultAppName),
		AppVersion: getEnv("APP_VERSION", defaultAppVersion),
		AppEnv:     getEnv("APP_ENV", "development"),

		HTTPAddr:        getEnv("HTTP_ADDR", defaultHTTPAddr),
		HTTPTimeout:     getDurationEnv("HTTP_TIMEOUT", defaultHTTPTimeout),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", defaultShutdownTimeout),

		DatabaseURL:      getEnv("DATABASE_URL", defaultDatabaseURL),
		DatabaseMaxConns: int32(getIntEnv("DATABASE_MAX_CONNS", int(defaultDatabaseMaxConns))),
		DatabaseMinConns: int32(getIntEnv("DATABASE_MIN_CONNS", int(defaultDatabaseMinConns))),

		RedisAddr:     getEnv("REDIS_ADDR", defaultRedisAddr),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       getIntEnv("REDIS_DB", 0),

		DScanCacheTTL:     getDurationEnv("DSCAN_CACHE_TTL", defaultDScanCacheTTL),
		ESICharacterTTL:   getDurationEnv("ESI_CHARACTER_CACHE_TTL", defaultESICharacterTTL),
		ESIAffiliationTTL: getDurationEnv("ESI_AFFILIATION_CACHE_TTL", defaultESIAffiliationTTL),
		ESIEntityNameTTL:  getDurationEnv("ESI_ENTITY_NAME_CACHE_TTL", defaultESIEntityNameTTL),

		ESIBaseURL:      getEnv("ESI_BASE_URL", defaultESIBaseURL),
		ESIDatasource:   getEnv("ESI_DATASOURCE", defaultESIDatasource),
		ESILanguage:     getEnv("ESI_LANGUAGE", defaultESILanguage),
		ESIMaxRetries:   getIntEnv("ESI_MAX_RETRIES", defaultESIMaxRetries),
		ESIRetryDelay:   getDurationEnv("ESI_RETRY_DELAY", defaultESIRetryDelay),
		ESIIDsBatchSize: getIntEnv("ESI_IDS_BATCH_SIZE", defaultESIIDsBatchSize),
		ESIBatchSize:    getIntEnv("ESI_BATCH_SIZE", defaultESIBatchSize),
		SDEURL:          getEnv("SDE_URL", defaultSDEURL),

		ShortLinkLength: getIntEnv("SHORT_LINK_LENGTH", defaultShortLinkLength),
	}
}

func loadDotenvFiles() {
	if os.Getenv("DISABLE_DOTENV") == "1" {
		return
	}

	envPath := findUp(".env")
	if envPath != "" {
		_ = godotenv.Load(envPath)
	}

	localEnvPath := findUp(".env.local")
	if localEnvPath != "" {
		_ = godotenv.Overload(localEnvPath)
	}
}

func findUp(filename string) string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}
