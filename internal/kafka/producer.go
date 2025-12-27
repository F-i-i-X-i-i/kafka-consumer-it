package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"

	"kafka-consumer/internal/models"
)

// Producer represents a Kafka producer for sending image commands
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
	}
}

// SendMessage sends an image command to Kafka
func (p *Producer) SendMessage(ctx context.Context, cmd models.ImageCommand) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(cmd.ID),
		Value: data,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("[PRODUCER] Ошибка отправки сообщения: %v", err)
		return err
	}

	log.Printf("[PRODUCER] Сообщение отправлено: ID=%s, Command=%s", cmd.ID, cmd.Command)
	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
