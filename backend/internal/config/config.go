package config

import "os"

type Config struct {
	Port        string
	JWTSecret   string
	DBPath      string
	FrontendURL string
}

func Load() Config {
	return Config{
		Port:        envOrDefault("PORT", "8080"),
		JWTSecret:   envOrDefault("JWT_SECRET", "supersecret"),
		DBPath:      envOrDefault("DB_PATH", "./data/app.db"),
		FrontendURL: envOrDefault("FRONTEND_URL", "http://localhost:5173"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
