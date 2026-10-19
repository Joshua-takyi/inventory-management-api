package config

import "os"

type Config struct {
	DbUrl string
	Port  string
}

func NewConfig() (*Config, error) {
	return &Config{
		DbUrl: fallbackEnv("DB_URL", ""),
		Port:  fallbackEnv("PORT", ""),
	}, nil

}

func fallbackEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
