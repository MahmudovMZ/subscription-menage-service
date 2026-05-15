package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env  string
	Port string
	DB   DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func MustLoad() *Config {

	log.Println("Loading configuration...")

	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	cfg := &Config{
		Env:  getEnvOrPanic("ENV"),
		Port: getEnvOrPanic("SERVER_PORT"),
		DB: DBConfig{
			Host:     getEnvOrPanic("DB_HOST"),
			Port:     getEnvOrPanic("DB_PORT"),
			User:     getEnvOrPanic("DB_USER"),
			Password: getEnvOrPanic("DB_PASSWORD"),
			Name:     getEnvOrPanic("DB_NAME"),
			SSLMode:  getEnvOrPanic("DB_SSLMODE"),
		},
	}

	log.Printf("Configuration loaded successfully. Server will run on port: %s", cfg.Port)
	return cfg
}

func getEnvOrPanic(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("Critical error: environment variable %s is not set", key))
	}
	return value
}
