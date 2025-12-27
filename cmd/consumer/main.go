package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kafka-consumer/internal/api"
	"kafka-consumer/internal/config"
	"kafka-consumer/internal/kafka"
	"kafka-consumer/internal/processor"
)

func main() {
	log.Println("=== Kafka Consumer для обработки AI-команд ===")

	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Конфигурация: brokers=%v, topic=%s, groupID=%s",
		cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)

	// Create API server for health checks
	apiServer := api.NewServer()

	// Create processor (stub for now)
	proc := processor.NewStubProcessor()

	// Create Kafka consumer
	consumer := kafka.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		proc,
	)
	consumer.SetOnMessageProcessed(apiServer.IncrementMessagesCount)

	// Create Kafka producer for sending test messages
	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	apiServer.SetProducer(producer)
	defer producer.Close()

	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Получен сигнал: %v, инициируем остановку...", sig)
		cancel()
	}()

	// Start HTTP server for health checks and message sending
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", apiServer.HealthHandler)
	mux.HandleFunc("/ready", apiServer.ReadyHandler)
	mux.HandleFunc("/stats", apiServer.StatsHandler)
	mux.HandleFunc("/send", apiServer.SendHandler)

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP сервер запущен на порту %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка HTTP сервера: %v", err)
		}
	}()

	// Start Kafka consumer in a goroutine
	go func() {
		log.Println("Запуск Kafka consumer...")
		apiServer.SetKafkaConnected(true)
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Ошибка consumer: %v", err)
			apiServer.SetKafkaConnected(false)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown of HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке HTTP сервера: %v", err)
	}

	log.Println("Consumer успешно остановлен")
}
