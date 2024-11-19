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

type SMTP struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Env    string       `mapstructure:"environment"`
	Server ServerConfig `mapstructure:"server"`
	SMTP   SMTP         `mapstructure:"smtp"`
}

// LoadConfig initializes, validates, and returns the application configuration
func LoadConfig() (*Config, error) {
	// Initialize and set defaults
	if err := initializeViper(); err != nil {
		return nil, fmt.Errorf("failed to initialize viper: %w", err)
	}

	// Unmarshal configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return &config, nil
}

func initializeViper() error {
	// Set up viper to read from both config files and environment variables
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")

	// Environment variable handling
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Override SMTP credentials from environment if available
	if username, password := viper.GetString("SMTP_USERNAME"), viper.GetString("SMTP_PASSWORD"); username != "" && password != "" {
		viper.Set("smtp.username", username)
		viper.Set("smtp.password", password)
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("app.name", "doozip")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("environment", "development")

	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.shutdown_timeout", "5s")
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")

	viper.SetDefault("smtp.host", "smtp.example.com")
	viper.SetDefault("smtp.port", "587")
}

func validateConfig(config *Config) error {
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
	if config.Server.ShutdownTimeout <= 0 || config.Server.ReadTimeout <= 0 || config.Server.WriteTimeout <= 0 || config.Server.IdleTimeout <= 0 {
		return fmt.Errorf("all server timeouts must be positive")
	}
	return nil
}

func isValidEnvironment(env string) bool {
	validEnvs := map[string]struct{}{
		"development": {},
		"production":  {},
	}
	_, valid := validEnvs[env]
	return valid
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
	SMTP Host:             %s
	SMTP Port:             %s
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
		c.SMTP.Host,
		c.SMTP.Port,
	)
}

// GetAddress returns the full address string for the server
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
