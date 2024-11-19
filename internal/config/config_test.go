package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name         string
		configFile   string
		envVars      map[string]string
		expectedErr  bool
		validateFunc func(*testing.T, *Config)
	}{
		{
			name: "Valid configuration",
			configFile: `
app:
  name: "testapp"
  version: "1.0.0"
environment: "development"
server:
  host: "localhost"
  port: 8080
  shutdown_timeout: "5s"
  read_timeout: "5s"
  write_timeout: "10s"
  idle_timeout: "60s"
smtp:
  host: "smtp.test.com"
  port: "587"
`,
			envVars:     map[string]string{},
			expectedErr: false,
			validateFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "testapp", cfg.App.Name)
				assert.Equal(t, "1.0.0", cfg.App.Version)
				assert.Equal(t, "development", cfg.Env)
				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, "smtp.test.com", cfg.SMTP.Host)
			},
		},
		{
			name: "Override SMTP credentials with env vars",
			configFile: `
app:
  name: "testapp"
  version: "1.0.0"
environment: "development"
server:
  port: 8080
smtp:
  host: "smtp.test.com"
  port: "587"
`,
			envVars: map[string]string{
				"SMTP_USERNAME": "test@example.com",
				"SMTP_PASSWORD": "secretpassword",
			},
			expectedErr: false,
			validateFunc: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "test@example.com", cfg.SMTP.Username)
				assert.Equal(t, "secretpassword", cfg.SMTP.Password)
			},
		},
		{
			name: "Invalid port",
			configFile: `
app:
  name: "testapp"
  version: "1.0.0"
environment: "development"
server:
  port: 70000
`,
			envVars:      map[string]string{},
			expectedErr:  true,
			validateFunc: nil,
		},
		{
			name: "Invalid environment",
			configFile: `
app:
  name: "testapp"
  version: "1.0.0"
environment: "invalid"
server:
  port: 8080
`,
			envVars:      map[string]string{},
			expectedErr:  true,
			validateFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setupTest(t, tt.configFile, tt.envVars)
			defer cleanupTest(t)

			// Test
			cfg, err := LoadConfig()

			if tt.expectedErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.validateFunc != nil {
				tt.validateFunc(t, cfg)
			}
		})
	}
}

func TestConfig_GetAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	assert.Equal(t, "localhost:8080", cfg.GetAddress())
}

func TestConfig_String(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Name:    "testapp",
			Version: "1.0.0",
		},
		Env: "development",
		Server: ServerConfig{
			Host:            "localhost",
			Port:            8080,
			ShutdownTimeout: 5 * time.Second,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     60 * time.Second,
		},
		SMTP: SMTP{
			Host: "smtp.test.com",
			Port: "587",
		},
	}

	str := cfg.String()
	assert.Contains(t, str, "testapp")
	assert.Contains(t, str, "1.0.0")
	assert.Contains(t, str, "development")
	assert.Contains(t, str, "localhost")
	assert.Contains(t, str, "8080")
}

// Helper functions for setting up and cleaning up tests
func setupTest(t *testing.T, configContent string, envVars map[string]string) {
	// Create temporary config file
	err := os.MkdirAll("./config", 0o755)
	require.NoError(t, err)

	err = os.WriteFile("./config/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Reset viper
	viper.Reset()
}

func cleanupTest(t *testing.T) {
	// Clean up config directory
	err := os.RemoveAll("./config")
	require.NoError(t, err)

	// Clear environment variables
	os.Clearenv()
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				App: AppConfig{
					Name:    "testapp",
					Version: "1.0.0",
				},
				Env: "development",
				Server: ServerConfig{
					Port:            8080,
					ShutdownTimeout: 5 * time.Second,
					ReadTimeout:     5 * time.Second,
					WriteTimeout:    10 * time.Second,
					IdleTimeout:     60 * time.Second,
				},
			},
			expectedErr: false,
		},
		{
			name: "Missing app name",
			config: &Config{
				App: AppConfig{
					Version: "1.0.0",
				},
				Env: "development",
				Server: ServerConfig{
					Port: 8080,
				},
			},
			expectedErr: true,
		},
		{
			name: "Invalid port",
			config: &Config{
				App: AppConfig{
					Name:    "testapp",
					Version: "1.0.0",
				},
				Env: "development",
				Server: ServerConfig{
					Port: 70000,
				},
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
