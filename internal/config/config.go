package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Logging    LoggingConfig
	Crawler    CrawlerConfig
	Browser    BrowserConfig
	Proxy      ProxyConfig
	Auth       AuthConfig
	Embroidery EmbroideryConfig
}

type ServerConfig struct {
	Host      string
	Port      int
	APIPrefix string
	CORSOrigin string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type CrawlerConfig struct {
	MaxWorkers            int
	RateLimitPerSecond    int
	RequestTimeout        time.Duration
	RetryMaxAttempts      int
	RetryBackoffMultiplier int
}

type BrowserConfig struct {
	Headless      bool
	Timeout       time.Duration
	UserDataDir   string
}

type ProxyConfig struct {
	Enabled                bool
	HealthCheckInterval    time.Duration
	MaxFailures            int
}

type AuthConfig struct {
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
	AdminTokenLifetime time.Duration
	RateLimitRequests int // Number of requests per window
	RateLimitWindow    time.Duration // Time window for rate limiting
}

type EmbroideryConfig struct {
	BaseURL       string
	AuthToken     string
	Cookies       string
	PageSize      int
	CheckInterval time.Duration
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")
	
	// Set defaults
	setDefaults()
	
	// Read from environment variables
	viper.AutomaticEnv()
	
	// Try to read .env file (optional)
	_ = viper.ReadInConfig()
	
	config := &Config{
		Server: ServerConfig{
			Host:      viper.GetString("SERVER_HOST"),
			Port:      viper.GetInt("SERVER_PORT"),
			APIPrefix: viper.GetString("API_PREFIX"),
			CORSOrigin: viper.GetString("CORS_ORIGIN"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetInt("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
		},
		Crawler: CrawlerConfig{
			MaxWorkers:            viper.GetInt("MAX_WORKERS"),
			RateLimitPerSecond:    viper.GetInt("RATE_LIMIT_PER_SECOND"),
			RequestTimeout:        parseDuration(viper.GetString("REQUEST_TIMEOUT")),
			RetryMaxAttempts:      viper.GetInt("RETRY_MAX_ATTEMPTS"),
			RetryBackoffMultiplier: viper.GetInt("RETRY_BACKOFF_MULTIPLIER"),
		},
		Browser: BrowserConfig{
			Headless:    viper.GetBool("HEADLESS"),
			Timeout:     parseDuration(viper.GetString("BROWSER_TIMEOUT")),
			UserDataDir: viper.GetString("USER_DATA_DIR"),
		},
		Proxy: ProxyConfig{
			Enabled:             viper.GetBool("PROXY_ENABLED"),
			HealthCheckInterval: parseDuration(viper.GetString("PROXY_HEALTH_CHECK_INTERVAL")),
			MaxFailures:         viper.GetInt("PROXY_MAX_FAILURES"),
		},
		Auth: AuthConfig{
			JWTSecret:         viper.GetString("JWT_SECRET"),
			JWTExpiration:     parseDuration(viper.GetString("JWT_EXPIRATION")),
			RefreshExpiration: parseDuration(viper.GetString("REFRESH_TOKEN_EXPIRATION")),
			AdminTokenLifetime: parseDuration(viper.GetString("ADMIN_TOKEN_LIFETIME")),
			RateLimitRequests: viper.GetInt("AUTH_RATE_LIMIT_REQUESTS"),
			RateLimitWindow:    parseDuration(viper.GetString("AUTH_RATE_LIMIT_WINDOW")),
		},
		Embroidery: EmbroideryConfig{
			BaseURL:       viper.GetString("EMBROIDERY_BASE_URL"),
			AuthToken:     viper.GetString("EMBROIDERY_AUTH_TOKEN"),
			Cookies:       viper.GetString("EMBROIDERY_COOKIES"),
			PageSize:      viper.GetInt("EMBROIDERY_PAGE_SIZE"),
			CheckInterval: parseDuration(viper.GetString("EMBROIDERY_CHECK_INTERVAL")),
		},
	}
	
	// Validate required configuration values
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	
	return config, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	if cfg.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	return nil
}

func setDefaults() {
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("SERVER_PORT", 8009)
	viper.SetDefault("API_PREFIX", "/api/v1")
	viper.SetDefault("CORS_ORIGIN", "*") // Default to * for development, should be set to specific origin in production
	
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "crawler")
	// DB_PASSWORD must be set via environment variable - no default for security
	viper.SetDefault("DB_NAME", "crawler_db")
	viper.SetDefault("DB_SSLMODE", "disable")
	
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	
	viper.SetDefault("MAX_WORKERS", 10)
	viper.SetDefault("RATE_LIMIT_PER_SECOND", 5)
	viper.SetDefault("REQUEST_TIMEOUT", "30s")
	viper.SetDefault("RETRY_MAX_ATTEMPTS", 3)
	viper.SetDefault("RETRY_BACKOFF_MULTIPLIER", 2)
	
	viper.SetDefault("HEADLESS", true)
	viper.SetDefault("BROWSER_TIMEOUT", "60s")
	viper.SetDefault("USER_DATA_DIR", "/tmp/browser-data")
	
	viper.SetDefault("PROXY_ENABLED", true)
	viper.SetDefault("PROXY_HEALTH_CHECK_INTERVAL", "5m")
	viper.SetDefault("PROXY_MAX_FAILURES", 3)
	
	// JWT_SECRET must be set via environment variable - no default for security
	viper.SetDefault("JWT_EXPIRATION", "24h")
	viper.SetDefault("REFRESH_TOKEN_EXPIRATION", "168h") // 7 days
	viper.SetDefault("ADMIN_TOKEN_LIFETIME", "8760h")    // 1 year
	viper.SetDefault("AUTH_RATE_LIMIT_REQUESTS", 5)      // 5 requests per window
	viper.SetDefault("AUTH_RATE_LIMIT_WINDOW", "15m")   // 15 minute window
	
	viper.SetDefault("EMBROIDERY_BASE_URL", "https://www.embroiderydesigns.com/es/prdsrch")
	// EMBROIDERY_AUTH_TOKEN must be set via environment variable - no default for security
	viper.SetDefault("EMBROIDERY_COOKIES", "")
	viper.SetDefault("EMBROIDERY_PAGE_SIZE", 120)
	viper.SetDefault("EMBROIDERY_CHECK_INTERVAL", "6h")
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

