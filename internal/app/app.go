package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kafka-consumer/internal/config"
	"kafka-consumer/internal/delivery/api"
	"kafka-consumer/internal/delivery/queue"
	"kafka-consumer/internal/kafka"
	"kafka-consumer/internal/pkg/logger"
	"kafka-consumer/internal/pkg/metrics"
	"kafka-consumer/internal/processor"
)

// Application represents the main application
type Application struct {
	cfg        *config.Config
	apiServer  *api.Server
	handler    *queue.Handler
	producer   *kafka.Producer
	httpServer *http.Server
}

// New creates a new Application instance
func New(cfg *config.Config) *Application {
	return &Application{
		cfg: cfg,
	}
}

// Init initializes all application dependencies
// Panics if critical dependencies fail (Kafka connection)
func (a *Application) Init() error {
	// Initialize logger
	logger.Init(a.cfg.LogLevel)

	logger.Info("Starting Kafka Consumer for AI commands",
		"brokers", a.cfg.KafkaBrokers,
		"topic", a.cfg.KafkaTopic,
		"group_id", a.cfg.KafkaGroupID,
		"processor_mode", a.cfg.ProcessorMode,
		"message_format", a.cfg.MessageFormat,
	)

	// Create API server with chi router
	a.apiServer = api.NewServer()

	// Create processor based on configuration
	var proc processor.Processor
	var err error
	if a.cfg.ProcessorMode == "real" {
		logger.Info("Using real image processor", "output_dir", a.cfg.OutputDir)
		proc, err = processor.NewRealProcessor(a.cfg)
		if err != nil {
			logger.Error("Failed to create processor", "error", err)
			return err
		}
	} else {
		logger.Info("Using stub processor")
		proc = processor.NewStubProcessor()
	}

	// Create message decoder
	format := queue.FormatJSON
	if a.cfg.MessageFormat == "protobuf" {
		format = queue.FormatProtobuf
	}
	decoder := queue.NewDecoder(format)

	// Create Kafka queue handler (consumer)
	a.handler = queue.NewHandler(
		a.cfg.KafkaBrokers,
		a.cfg.KafkaTopic,
		a.cfg.KafkaGroupID,
		proc,
		decoder,
	)

	// Create Kafka producer for sending test messages
	a.producer = kafka.NewProducer(a.cfg.KafkaBrokers, a.cfg.KafkaTopic)
	if a.cfg.MessageFormat == "protobuf" {
		a.producer.SetMessageFormat(kafka.FormatProtobuf)
	}
	a.apiServer.SetProducer(a.producer)

	// Setup HTTP server with chi router
	mux := http.NewServeMux()
	mux.Handle("/", a.apiServer.Router())
	mux.Handle("/metrics", metrics.Handler())

	a.httpServer = &http.Server{
		Addr:    ":" + a.cfg.HTTPPort,
		Handler: mux,
	}

	return nil
}

// Run starts the application and blocks until shutdown
func (a *Application) Run(ctx context.Context) error {
	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	// Start HTTP server
	go func() {
		logger.Info("HTTP server started", "addr", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Start Kafka consumer handler
	errChan := make(chan error, 1)
	go func() {
		logger.Info("Starting Kafka consumer")
		if err := a.handler.Start(ctx); err != nil {
			errChan <- fmt.Errorf("kafka consumer failed: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
	case err := <-errChan:
		logger.Error("Consumer error", "error", err)
		return err
	}

	return a.Shutdown()
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown() error {
	logger.Info("Starting graceful shutdown")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	// Close producer
	if a.producer != nil {
		if err := a.producer.Close(); err != nil {
			logger.Error("Producer close error", "error", err)
		}
	}

	logger.Info("Application stopped successfully")
	return nil
}
