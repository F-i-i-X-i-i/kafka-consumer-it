package processor

import (
	"fmt"
	"log"

	"kafka-consumer/internal/models"
)

// Processor defines the interface for processing image commands
type Processor interface {
	Process(cmd models.ImageCommand) error
}

// StubProcessor is a stub implementation that logs commands for demonstration
type StubProcessor struct {
	ProcessedCommands []models.ImageCommand
}

// NewStubProcessor creates a new stub processor
func NewStubProcessor() *StubProcessor {
	return &StubProcessor{
		ProcessedCommands: make([]models.ImageCommand, 0),
	}
}

// Process logs the command and stores it for later inspection
func (p *StubProcessor) Process(cmd models.ImageCommand) error {
	log.Printf("[PROCESSOR] Получена команда: ID=%s, Type=%s, ImageURL=%s",
		cmd.ID, cmd.Command, cmd.ImageURL)

	// Validate command
	if cmd.ID == "" {
		return fmt.Errorf("command ID is required")
	}
	if cmd.ImageURL == "" {
		return fmt.Errorf("image URL is required")
	}

	// Store command for inspection (useful for testing)
	p.ProcessedCommands = append(p.ProcessedCommands, cmd)

	// Log processing based on command type
	switch cmd.Command {
	case models.CommandResize:
		log.Printf("[PROCESSOR] Изменение размера изображения: %v", cmd.Parameters)
	case models.CommandFilter:
		log.Printf("[PROCESSOR] Применение фильтра к изображению: %v", cmd.Parameters)
	case models.CommandTransform:
		log.Printf("[PROCESSOR] Трансформация изображения: %v", cmd.Parameters)
	case models.CommandAnalyze:
		log.Printf("[PROCESSOR] Анализ изображения ИИ: %v", cmd.Parameters)
	default:
		log.Printf("[PROCESSOR] Неизвестная команда: %s", cmd.Command)
	}

	log.Printf("[PROCESSOR] Команда %s успешно обработана (заглушка)", cmd.ID)
	return nil
}

// GetProcessedCount returns the number of processed commands
func (p *StubProcessor) GetProcessedCount() int {
	return len(p.ProcessedCommands)
}

// GetLastCommand returns the last processed command
func (p *StubProcessor) GetLastCommand() *models.ImageCommand {
	if len(p.ProcessedCommands) == 0 {
		return nil
	}
	return &p.ProcessedCommands[len(p.ProcessedCommands)-1]
}
