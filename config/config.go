package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	Env            string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	JWTExpiryHours string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: No .env file found")
	}
	return &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPassword:     getEnv("DB_PASSWORD", "root@123"),
		DBName:         getEnv("DB_NAME", "library_db"),
		JWTSecret:      getEnv("JWT_SECRET", "super_secret_jwt_key_library_app"),
		JWTExpiryHours: getEnv("JWT_EXPIRY_HOURS", "24"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
