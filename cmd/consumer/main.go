package main

import (
	"context"
	"log"

	"kafka-consumer/internal/app"
	"kafka-consumer/internal/config"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create and initialize application
	application := app.New(cfg)
	if err := application.Init(); err != nil {
		log.Fatalf("Ошибка инициализации приложения: %v", err)
	}

	// Run the application
	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Ошибка запуска приложения: %v", err)
	}
}
