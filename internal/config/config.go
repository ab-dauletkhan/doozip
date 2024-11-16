package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
}

type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Env    string       `mapstructure:"environment"`
	Server ServerConfig `mapstructure:"server"`
}

// LoadConfig loads and validates configuration from file and environment
func LoadConfig() (*Config, error) {
	// Initialize viper with defaults
	if err := initializeViper(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func initializeViper() error {
	// Set up viper to read from both config file and environment variables
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")

	// Enable viper to read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "doozip")
	viper.SetDefault("app.version", "1.0.0")

	// Environment default
	viper.SetDefault("environment", "development")

	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.shutdown_timeout", "5s")
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")
}

func validateConfig(config *Config) error {
	// Basic validation
	if config.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if config.App.Version == "" {
		return fmt.Errorf("app version is required")
	}
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	if !isValidEnvironment(config.Env) {
		return fmt.Errorf("invalid environment: %s", config.Env)
	}

	// Validate timeouts are positive
	if config.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}
	if config.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	if config.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	if config.Server.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	return nil
}

func isValidEnvironment(env string) bool {
	validEnvs := map[string]bool{
		"development": true,
		"production":  true,
	}
	return validEnvs[env]
}

// String returns a string representation of the config for debugging
func (c *Config) String() string {
	return fmt.Sprintf(
		`Config:
	App Name:              %s
	App Version:           %s
	Environment:           %s
	Server Host:           %s
	Server Port:           %d
	Shutdown Timeout:      %s
	Read Timeout:          %s
	Write Timeout:         %s
	Idling Timeout:        %s
	`,
		c.App.Name,
		c.App.Version,
		c.Env,
		c.Server.Host,
		c.Server.Port,
		c.Server.ShutdownTimeout,
		c.Server.ReadTimeout,
		c.Server.WriteTimeout,
		c.Server.IdleTimeout,
	)
}

// GetAddress returns the full address string for the server
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
