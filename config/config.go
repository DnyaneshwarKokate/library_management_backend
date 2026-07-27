package config

import (
	"os"
	"strconv"
)

// DatabaseConfig represents the MySQL database configuration.
type MySQLDatabaseConfig struct {
	Username     string
	Password     string
	Host         string
	Port         string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
}

func GetPrimaryMySQLDBConfig() MySQLDatabaseConfig {
	username := getEnv("DB_USER", getEnv("DB_USERNAME", "root"))
	password := getEnv("DB_PASSWORD", "root@123")
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	database := getEnv("DB_NAME", getEnv("DB_DATABASE_NAME", "library_db"))

	return MySQLDatabaseConfig{
		Username:     username,
		Password:     password,
		Host:         host,
		Port:         port,
		Database:     database,
		MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 5),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
