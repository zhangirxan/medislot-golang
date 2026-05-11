package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort               string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	DBSSLMode                string
	JWTSecret                string
	JWTExpiryHours           int
	WorkerIntervalSeconds    int
	AppointmentExpiryMinutes int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}
	workerInterval, err := strconv.Atoi(getEnv("WORKER_INTERVAL_SECONDS", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid WORKER_INTERVAL_SECONDS: %w", err)
	}
	apptExpiry, err := strconv.Atoi(getEnv("APPOINTMENT_EXPIRY_MINUTES", "30"))
	if err != nil {
		return nil, fmt.Errorf("invalid APPOINTMENT_EXPIRY_MINUTES: %w", err)
	}

	return &Config{
		ServerPort:               getEnv("SERVER_PORT", "8080"),
		DBHost:                   getEnv("DB_HOST", "localhost"),
		DBPort:                   getEnv("DB_PORT", "5432"),
		DBUser:                   getEnv("DB_USER", "postgres"),
		DBPassword:               getEnv("DB_PASSWORD", "postgres"),
		DBName:                   getEnv("DB_NAME", "medislot"),
		DBSSLMode:                getEnv("DB_SSLMODE", "disable"),
		JWTSecret:                getEnv("JWT_SECRET", "change-me"),
		JWTExpiryHours:           jwtExpiry,
		WorkerIntervalSeconds:    workerInterval,
		AppointmentExpiryMinutes: apptExpiry,
	}, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
