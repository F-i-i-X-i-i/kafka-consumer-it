package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"kafka-consumer/internal/models"
	"kafka-consumer/internal/processor"
)

// Consumer represents a Kafka consumer for image commands
type Consumer struct {
	reader             *kafka.Reader
	processor          processor.Processor
	onMessageProcessed func()
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string, proc processor.Processor) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		CommitInterval: time.Second,
	})

	return &Consumer{
		reader:    reader,
		processor: proc,
	}
}

// SetOnMessageProcessed sets a callback to be called after each message is processed
func (c *Consumer) SetOnMessageProcessed(callback func()) {
	c.onMessageProcessed = callback
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("[CONSUMER] Запуск consumer, подключение к топику: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("[CONSUMER] Получен сигнал остановки, завершение работы...")
			return c.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil // Context cancelled, graceful shutdown
				}
				log.Printf("[CONSUMER] Ошибка чтения сообщения: %v", err)
				continue
			}

			log.Printf("[CONSUMER] Получено сообщение: partition=%d, offset=%d, key=%s",
				msg.Partition, msg.Offset, string(msg.Key))

			// Deserialize message
			var cmd models.ImageCommand
			if err := json.Unmarshal(msg.Value, &cmd); err != nil {
				log.Printf("[CONSUMER] Ошибка десериализации JSON: %v", err)
				// Commit message even on error to avoid infinite retry
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("[CONSUMER] Ошибка коммита сообщения: %v", err)
				}
				continue
			}

			// Process the command
			if err := c.processor.Process(cmd); err != nil {
				log.Printf("[CONSUMER] Ошибка обработки команды %s: %v", cmd.ID, err)
			} else {
				// Call callback on successful processing
				if c.onMessageProcessed != nil {
					c.onMessageProcessed()
				}
			}

			// Commit the message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[CONSUMER] Ошибка коммита сообщения: %v", err)
			}
		}
	}
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	log.Println("[CONSUMER] Закрытие соединения с Kafka...")
	return c.reader.Close()
}
