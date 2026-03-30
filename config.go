package main

import (
	"github.com/caarlos0/env/v11"
)

// Config holds all runtime configuration, parsed from environment variables
// via github.com/caarlos0/env/v11.
type Config struct {
	// Server
	Addr string `env:"PROXY_ADDR" envDefault:":8080"`

	// Allowlist — if set, only these upstream hosts are permitted.
	// Comma-separated, e.g. "api.example.com,ipinfo.io"
	// If empty, all hosts are allowed (open proxy — use only in trusted environments).
	AllowedHosts []string `env:"PROXY_ALLOWED_HOSTS" envSeparator:","`

	// CORS
	AllowOrigins           string `env:"CORS_ALLOW_ORIGINS"  envDefault:"*"`
	AllowMethods           string `env:"CORS_ALLOW_METHODS"  envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	AllowHeaders           string `env:"CORS_ALLOW_HEADERS"  envDefault:"Accept,Authorization,Content-Type,X-Requested-With"`
	MaxAge                 string `env:"CORS_MAX_AGE"        envDefault:"86400"`
	AllowCredentials       bool   `env:"CORS_ALLOW_CREDENTIALS"`
	HideAllowOriginsHeader bool   `env:"CORS_HIDE_ALLOW_ORIGINS_HEADER"  envDefault:"false"`

	// Logging
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`
	LogLevel  string `env:"LOG_LEVEL"  envDefault:"info"`
}

func loadConfig() *Config {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		panic("config error: " + err.Error())
	}
	return &cfg
}
