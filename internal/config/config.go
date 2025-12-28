package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Kafka settings
	KafkaBrokers []string `mapstructure:"kafka_brokers"`
	KafkaTopic   string   `mapstructure:"kafka_topic"`
	KafkaGroupID string   `mapstructure:"kafka_group_id"`

	// Processor settings
	ProcessorMode string `mapstructure:"processor_mode"` // "stub" or "real"
	OutputDir     string `mapstructure:"output_dir"`     // Directory for processed images

	// Message format
	MessageFormat string `mapstructure:"message_format"` // "json" or "protobuf"

	// HTTP settings
	HTTPPort string `mapstructure:"http_port"`

	// Logging
	LogLevel string `mapstructure:"log_level"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig() *Config {
	v := viper.New()

	// Set defaults
	v.SetDefault("kafka_brokers", "localhost:9092")
	v.SetDefault("kafka_topic", "image-commands")
	v.SetDefault("kafka_group_id", "image-processor-group")
	v.SetDefault("processor_mode", "stub")
	v.SetDefault("output_dir", "/tmp/processed-images")
	v.SetDefault("message_format", "json")
	v.SetDefault("http_port", "8080")
	v.SetDefault("log_level", "info")

	// Read from environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read from config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/kafka-consumer")

	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	}

	cfg := &Config{}

	// Manually bind because Viper env binding requires explicit mapping
	cfg.KafkaBrokers = parseBrokers(v.GetString("kafka_brokers"))
	cfg.KafkaTopic = v.GetString("kafka_topic")
	cfg.KafkaGroupID = v.GetString("kafka_group_id")
	cfg.ProcessorMode = v.GetString("processor_mode")
	cfg.OutputDir = v.GetString("output_dir")
	cfg.MessageFormat = v.GetString("message_format")
	cfg.HTTPPort = v.GetString("http_port")
	cfg.LogLevel = v.GetString("log_level")

	return cfg
}

// parseBrokers parses comma-separated broker list
func parseBrokers(brokers string) []string {
	if brokers == "" {
		return []string{"localhost:9092"}
	}
	return strings.Split(brokers, ",")
}
