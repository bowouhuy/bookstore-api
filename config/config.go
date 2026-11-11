package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type (
	Config struct {
		App      App
		MasterDB DB
		SlaveDB  DB
		JWT      JWT
		Ws       Websocket
	}

	App struct {
		Name         string
		Env          string
		Port         int
		DefaultLang  string
		LogLevel     string
		ReadTimeout  int
		WriteTimeout int
	}

	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		DB       string
		SSLMode  string
	}

	JWT struct {
		Secret string
		TTL    time.Duration
	}

	Websocket struct {
		Host string
		Port int

		RoomIDLength      int
		MaxCachedMessages int
		MaxMessageLength  int
		Timeout           int
		RateLimitInterval int
		RateLimitMessages int
		MaxRooms          int
		MaxPeerPerRoom    int
		PeerHandleFormat  string
		RoomTimeout       int
		RoomAge           int
	}
)

func New() *Config {
	return &Config{
		App: App{
			Name:         getEnv("APP_NAME", "bookstore-api"),
			Env:          getEnv("APP_ENV", "development"),
			Port:         getEnvAsInt("APP_PORT", 8080),
			DefaultLang:  getEnv("APP_DEFAULT_LANG", "en"),
			LogLevel:     getEnv("APP_LOG_LEVEL", "info"),
			ReadTimeout:  getEnvAsInt("APP_READ_TIMEOUT", 10),
			WriteTimeout: getEnvAsInt("APP_WRITE_TIMEOUT", 15),
		},
		MasterDB: DB{
			Host:     getEnv("POSTGRES_HOST_MASTER", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT_MASTER", 5432),
			User:     getEnv("POSTGRES_USER_MASTER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD_MASTER", "postgres"),
			DB:       getEnv("POSTGRES_DB_MASTER", "postgres"),
			SSLMode:  getEnv("POSTGRES_SSLMODE_MASTER", "disable"),
		},
		SlaveDB: DB{
			Host:     getEnv("POSTGRES_HOST_SLAVE", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT_SLAVE", 5432),
			User:     getEnv("POSTGRES_USER_SLAVE", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD_SLAVE", "postgres"),
			DB:       getEnv("POSTGRES_DB_SLAVE", "postgres"),
			SSLMode:  getEnv("POSTGRES_SSLMODE_SLAVE", "disable"),
		},
		JWT: JWT{
			Secret: getEnv("JWT_SECRET", ""),
			TTL:    getEnvAsDuration("JWT_TTL", 24*time.Hour),
		},
		Ws: Websocket{
			Host:              getEnv("WS_HOST", "localhost"),
			Port:              getEnvAsInt("WS_PORT", 3000),
			RoomIDLength:      getEnvAsInt("WS_ROOM_ID_LENGTH", 8),
			MaxCachedMessages: getEnvAsInt("WS_MAX_CACHED_MESSAGES", 1000),
			MaxMessageLength:  getEnvAsInt("WS_MAX_MESSAGE_LENGTH", 3000),
			Timeout:           getEnvAsInt("WS_TIMEOUT", 3),
			RateLimitInterval: getEnvAsInt("WS_RATE_LIMIT_INTERVAL", 3),
			RateLimitMessages: getEnvAsInt("WS_RATE_LIMIT_MESSAGES", 25),
			MaxRooms:          getEnvAsInt("WS_MAX_ROOMS", 100),
			MaxPeerPerRoom:    getEnvAsInt("WS_MAX_PEER_PER_ROOM", 100),
			PeerHandleFormat:  getEnv("WS_PEER_HANDLE_FORMAT", "Guest-%d"),
			RoomTimeout:       getEnvAsInt("WS_ROOM_TIMEOUT", 10),
			RoomAge:           getEnvAsInt("WS_ROOM_AGE", 24),
		},
	}
}

func (c *Config) Validate() error {
	var errs []error

	if c.App.Port < 1 || c.App.Port > 65535 {
		errs = append(errs, errors.New("APP_PORT must be between 1 and 65535"))
	}
	if c.JWT.Secret == "" {
		errs = append(errs, errors.New("JWT_SECRET is required"))
	} else if len(c.JWT.Secret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 characters"))
	}
	if c.JWT.TTL <= 0 {
		errs = append(errs, errors.New("JWT_TTL must be a positive duration such as 24h"))
	}

	return errors.Join(errs...)
}

func getEnv(key string, defaultVal string) string {
	if value, exist := os.LookupEnv(key); exist && value != "" {
		return value
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// getEnvAsDuration returns 0 for an unparsable value so Validate reports it
// instead of silently falling back to the default.
func getEnvAsDuration(name string, defaultVal time.Duration) time.Duration {
	valueStr := getEnv(name, "")
	if valueStr == "" {
		return defaultVal
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return 0
	}

	return value
}
