package queue

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"

	"kafka-consumer/internal/pkg/logger"
	"kafka-consumer/internal/processor"
)

// MessageDecoder decodes raw message bytes into commands
type MessageDecoder interface {
	DecodeCommand(data []byte) (interface{}, error)
}

// Handler handles Kafka messages
type Handler struct {
	reader    *kafka.Reader
	processor processor.Processor
	decoder   MessageDecoder
}

// NewHandler creates a new Kafka message handler
func NewHandler(brokers []string, topic, groupID string, proc processor.Processor, decoder MessageDecoder) *Handler {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		CommitInterval: time.Second,
	})

	return &Handler{
		reader:    reader,
		processor: proc,
		decoder:   decoder,
	}
}

// Start begins consuming messages from Kafka
// This method blocks until context is cancelled or an error occurs
func (h *Handler) Start(ctx context.Context) error {
	logger.Info("Starting Kafka consumer", "topic", h.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Received stop signal, shutting down consumer")
			return h.Close()
		default:
			if err := h.processNextMessage(ctx); err != nil {
				if ctx.Err() != nil {
					return nil // Context cancelled, graceful shutdown
				}
				// Log and continue on non-fatal errors
				logger.Error("Error processing message", "error", err)
			}
		}
	}
}

// processNextMessage fetches and processes a single message
func (h *Handler) processNextMessage(ctx context.Context) error {
	msg, err := h.reader.FetchMessage(ctx)
	if err != nil {
		return err
	}

	log := logger.With(
		"partition", msg.Partition,
		"offset", msg.Offset,
		"key", string(msg.Key),
	)

	log.Debug("Received message")

	// Decode message
	cmd, err := h.decoder.DecodeCommand(msg.Value)
	if err != nil {
		log.Error("Message decode error", "error", err)
		// Commit message even on error to avoid infinite retry
		return h.reader.CommitMessages(ctx, msg)
	}

	// Process the command - the processor handles its own metrics
	if err := h.processor.ProcessAny(ctx, cmd); err != nil {
		log.Error("Command processing error", "error", err)
	}

	// Commit the message
	return h.reader.CommitMessages(ctx, msg)
}

// Close closes the Kafka consumer
func (h *Handler) Close() error {
	logger.Info("Closing Kafka connection")
	return h.reader.Close()
}
