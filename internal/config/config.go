package config

import (
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "image-commands"
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "image-processor-group"
	}

	return &Config{
		KafkaBrokers: strings.Split(brokers, ","),
		KafkaTopic:   topic,
		KafkaGroupID: groupID,
	}
}
