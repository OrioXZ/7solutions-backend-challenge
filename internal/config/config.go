package config

import (
	"os"
	"strings"
)

const (
	defaultHTTPPort       = "8080"
	defaultMongoURI       = "mongodb://localhost:27017"
	defaultMongoDatabase  = "seven_solutions"
)

type Config struct {
	HTTPAddr     string
	MongoURI     string
	MongoDatabase string
}

func Load() Config {
	port := envOrDefault("HTTP_PORT", defaultHTTPPort)

	return Config{
		HTTPAddr:       ":" + port,
		MongoURI:       envOrDefault("MONGODB_URI", defaultMongoURI),
		MongoDatabase:  envOrDefault("MONGODB_DATABASE", defaultMongoDatabase),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
