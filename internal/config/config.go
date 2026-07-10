package config

import (
	"os"
	"strings"
)

const defaultHTTPPort = "8080"

type Config struct {
	HTTPAddr string
}

func Load() Config {
	port := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	if port == "" {
		port = defaultHTTPPort
	}

	return Config{
		HTTPAddr: ":" + port,
	}
}
