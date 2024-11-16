package config

import (
	"log"
	"path/filepath"
	"testing"

	"doozip/internal/utils"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestViper(t *testing.T) {
	cfg := viper.New()
	assert.NotNil(t, cfg)
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(utils.GetProjectRoot(), "config"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}
}

func TestInitConfig(t *testing.T) {
	// Set up a temporary configuration for testing
	viper.SetConfigType("yaml")   // or "json", "toml", etc. depending on your config format
	viper.SetConfigName("config") // name of your config file (without extension)
	viper.AddConfigPath(".")      // path to look for the config file in

	// Mock the configuration values for the test
	viper.Set("app.name", "MyApp")
	viper.Set("app.version", "1.0.0")
	viper.Set("environment", "development")
	viper.Set("server.host", "localhost")
	viper.Set("server.port", 8080)

	// Initialize configuration
	initConfig()

	// Retrieve values from viper
	appName := viper.GetString("app.name")
	appVersion := viper.GetString("app.version")
	env := viper.GetString("environment")
	host := viper.GetString("server.host")
	port := viper.GetInt("server.port")

	// Assertions to check if the values are as expected
	assert.Equal(t, "MyApp", appName, "App name should be 'MyApp'")
	assert.Equal(t, "1.0.0", appVersion, "App version should be '1.0.0'")
	assert.Equal(t, "development", env, "Environment should be 'development'")
	assert.Equal(t, "localhost", host, "Server host should be 'localhost'")
	assert.Equal(t, 8080, port, "Server port should be 8080")
}
