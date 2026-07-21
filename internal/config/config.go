package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// === BANCO DE DADOS ===
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// === SERVIDOR ===
	ServerPort string

	// === JWT ===
	JWTSecret     string
	JWTExpiresIn  time.Duration

	// === AMBIENTE ===
	Env string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Aviso: Arquivo .env não encontrado, usando variáveis de ambiente")
	}

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "cannacare"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// JWT
		JWTSecret:    getEnv("JWT_SECRET", "cannacare-super-secret-key-2026"),
		JWTExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),

		// Environment
		Env: getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}