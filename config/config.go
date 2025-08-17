package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	cfg  Config
	once sync.Once
)

type Config struct {
	appEnv  string
	appPort string

	dbHost string
	dbPort string
	dbUser string
	dbPass string
	dbName string
}

func (c Config) AppEnv() string {
	return c.appEnv
}

func (c Config) AppPort() string {
	return c.appPort
}

func (c Config) DbHost() string {
	return c.dbHost
}

func (c Config) DbPort() string {
	return c.dbPort
}

func (c Config) DbUser() string {
	return c.dbUser
}

func (c Config) DbPass() string {
	return c.dbPass
}

func (c Config) DbName() string {
	return c.dbName
}

func Load() Config {
	once.Do(func() {
		appEnv := optional("APP_ENV", "development")
		appPort := optional("APP_PORT", "8080")

		cfg = Config{
			appEnv:  appEnv,
			appPort: appPort,
			dbHost:  optional(fmt.Sprintf("%s_DB_HOST", strings.ToUpper(appEnv)), "clickhouse-dev"),
			dbPort:  optional(fmt.Sprintf("%s_DB_PORT", strings.ToUpper(appEnv)), "9000"),
			dbUser:  optional(fmt.Sprintf("%s_DB_USER", strings.ToUpper(appEnv)), "default"),
			dbPass:  optional(fmt.Sprintf("%s_DB_PASSWORD", strings.ToUpper(appEnv)), "password"),
			dbName:  optional(fmt.Sprintf("%s_DB_NAME", strings.ToUpper(appEnv)), "analyzify"),
		}
	})
	return cfg
}

func optional[T any](key string, defaultValue T) T {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return defaultValue
	}

	var result any

	switch any(defaultValue).(type) {
	case string:
		result = val
	case int:
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return defaultValue
		}
		result = parsed
	case bool:
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return defaultValue
		}
		result = parsed
	default:
		// unsupported type
		return defaultValue
	}

	return result.(T)
}
