package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"strings"
)

type Config struct {
	DBSQlitePath string
	DBDriver     string
	DBHost       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBPort       string
	RedisURL     string
	JWTSecret    string
	Port         string
	Environment  string
}

func Load() *Config {
	config := &Config{
		DBDriver:    getEnv("DB_DRIVER", "sqlite"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBUser:      getEnv("DB_USER", "admin"),
		DBPassword:  getDBPassword("password"),
		DBName:      getEnv("DB_NAME", "store"),
		DBPort:      getEnv("DB_PORT", "5432"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:   getJWTSecret(),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "dev"),
	}

	validateConfig(config)

	return config
}

func getDBPassword(defaultValue string) string {
	secretPath := "/run/secrets/db_password"
	if _, err := os.Stat(secretPath); err == nil {
		secretBytes, err := os.ReadFile(secretPath)
		if err != nil {
			log.Fatalf("Failed to read password file: %v", err)
		}
		passwd := strings.TrimSpace(string(secretBytes))
		log.Println("Database password loaded from Docker secret file.")
		return passwd
	}

	log.Println("Loading database password from enviroment variable")
	if passwd := os.Getenv("DB_PASSWORD"); passwd != "" {
		log.Println("Loaded from DB_PASSWORD")
		return passwd
	}

	log.Println("Using default value")
	return defaultValue
}

func getJWTSecret() string {
	secretPath := "/run/secrets/jwt_secret"
	if _, err := os.Stat(secretPath); err == nil {
		secretBytes, err := os.ReadFile(secretPath)
		if err != nil {
			log.Fatalf("Failed to read secret file: %v", err)
		}
		secret := strings.TrimSpace(string(secretBytes))
		log.Println("JWT Secret loaded from Docker secret file.")
		return secret
	}
	log.Println("JWT Secret loaded from environment variable (fallback).")

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		if len(secret) < 32 {
			log.Fatal("JWT_SECRET must be atlest 32 characters long")

		}
		return secret
	}

	if os.Getenv("ENVIRONMENT") == "prod" {
		log.Fatal("JWT_SECRET is required on production")
	}

	log.Println("Generating a random JWT_SECRET in dev environment")
	secret := generateRandomSecret()
	log.Printf("export JWT_SECRET=%s", secret)

	return secret
}

func generateRandomSecret() string {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Error generating random secret:", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func validateConfig(config *Config) {
	if len(config.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET is less than 32 chars")
	}

	if config.Environment == "prod" {
		if config.DBPassword == "password" {
			log.Fatal("Default database password is not allowed in production")
		}

		if config.JWTSecret == "super-secret-much-secure-very-wow" {
			log.Fatal("Default JWT_SECRET is not allowd in production")
		}

		log.Println("Production configuration validated")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
