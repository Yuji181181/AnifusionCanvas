package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	AppEnv            string
	DatabaseURL       string
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2PublicBaseURL   string
	R2EndpointURL     string
	R2Region          string
	ReplicateAPIToken string
	FrontendOrigin    string
	StudioStore       string
}

func Load() (Config, error) {
	loadDotEnv()

	cfg := Config{
		Port:              getEnv("API_PORT", "8080"),
		AppEnv:            getEnv("APP_ENV", "development"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		R2AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:          os.Getenv("R2_BUCKET"),
		R2PublicBaseURL:   os.Getenv("R2_PUBLIC_BASE_URL"),
		R2EndpointURL:     os.Getenv("R2_ENDPOINT_URL"),
		R2Region:          getEnv("R2_REGION", "auto"),
		ReplicateAPIToken: os.Getenv("REPLICATE_API_TOKEN"),
		FrontendOrigin:    getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		StudioStore:       getEnv("STUDIO_STORE", "memory"),
	}

	if cfg.AppEnv == "production" && cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.StudioStore == "database" && cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required when STUDIO_STORE=database")
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func loadDotEnv() {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}

	for _, candidate := range candidates {
		if err := godotenv.Load(candidate); err == nil {
			return
		}
	}
}
