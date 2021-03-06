package config

import (
	"os"
	"strings"

	"github.com/CESARBR/knot-babeltower/pkg/logging"

	"github.com/spf13/viper"
)

// Server represents the server configuration properties
type Server struct {
	Port int
}

// Logger represents the logger configuration properties
type Logger struct {
	Level string
}

// Users represents the users service to proxy request
type Users struct {
	Hostname string
	Port     uint16
}

// RabbitMQ represents the rabbitmq configuration properties
type RabbitMQ struct {
	URL string
}

// Things represents the things service to proxy request
type Things struct {
	Hostname string
	Port     uint16
}

// Config represents the service configuration
type Config struct {
	Server
	Logger
	Users
	RabbitMQ
	Things
}

func readFile(name string) {
	logger := logging.NewLogrus("error").Get("Config")
	viper.SetConfigName(name)
	if err := viper.ReadInConfig(); err != nil {
		logger.Fatalf("error reading config file, %s", err)
	}
}

// Load returns the service configuration
func Load() Config {
	var configuration Config
	logger := logging.NewLogrus("error").Get("Config")
	viper.AddConfigPath("internal/config")
	viper.SetConfigType("yaml")

	readFile("default")

	if os.Getenv("ENV") == "development" {
		readFile("development")
		if err := viper.MergeInConfig(); err != nil {
			logger.Fatalf("error reading config file, %s", err)
		}
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&configuration); err != nil {
		logger.Fatalf("error unmarshalling configuration, %s", err)
	}

	return configuration
}
