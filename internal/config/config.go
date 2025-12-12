package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RabbitMQConfig struct {
	URL        string
	Exchange   string
	Queue      string
	RoutingKey string
}

func Load() *Config {
	// Tenta carregar .env se existir
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return nil
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "orders_user"),
			Password: getEnv("DB_PASSWORD", "orders_pass"),
			DBName:   getEnv("DB_NAME", "orders_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:        getEnv("RABBITMQ_URL", "amqp://orders_user:orders_pass@localhost:5672/"),
			Exchange:   getEnv("RABBITMQ_EXCHANGE", "orders_exchange"),
			Queue:      getEnv("RABBITMQ_QUEUE", "order_events"),
			RoutingKey: getEnv("RABBITMQ_ROUTING_KEY", "order.created"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
