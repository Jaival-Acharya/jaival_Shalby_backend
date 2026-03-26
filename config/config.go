package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
	JWTExpiry  int
	Port       string
}

func Load() *Config {
	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		Port:       os.Getenv("PORT"),
	}

	// Parse JWT expiry hours
	jwtExpiry := os.Getenv("JWT_EXPIRY_HOURS")
	if jwtExpiry == "" {
		cfg.JWTExpiry = 24
	} else {
		hours, err := strconv.Atoi(jwtExpiry)
		if err != nil {
			log.Fatal("Invalid JWT_EXPIRY_HOURS:", err)
		}
		cfg.JWTExpiry = hours
	}

	// Validate required fields
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		log.Fatal("Missing required database configuration in .env file")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET not set in .env file")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg
}
