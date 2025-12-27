package processor

import (
	"testing"

	"kafka-consumer/internal/models"
)

func TestStubProcessor_Process_Success(t *testing.T) {
	processor := NewStubProcessor()

	cmd := models.ImageCommand{
		ID:       "test-123",
		Command:  models.CommandResize,
		ImageURL: "https://example.com/image.jpg",
		Parameters: map[string]interface{}{
			"width":  100,
			"height": 100,
		},
	}

	err := processor.Process(cmd)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if processor.GetProcessedCount() != 1 {
		t.Errorf("Expected 1 processed command, got: %d", processor.GetProcessedCount())
	}

	lastCmd := processor.GetLastCommand()
	if lastCmd == nil {
		t.Fatal("Expected last command to be not nil")
	}
	if lastCmd.ID != cmd.ID {
		t.Errorf("Expected ID %s, got: %s", cmd.ID, lastCmd.ID)
	}
}

func TestStubProcessor_Process_EmptyID(t *testing.T) {
	processor := NewStubProcessor()

	cmd := models.ImageCommand{
		ID:       "",
		Command:  models.CommandResize,
		ImageURL: "https://example.com/image.jpg",
	}

	err := processor.Process(cmd)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestStubProcessor_Process_EmptyImageURL(t *testing.T) {
	processor := NewStubProcessor()

	cmd := models.ImageCommand{
		ID:       "test-123",
		Command:  models.CommandResize,
		ImageURL: "",
	}

	err := processor.Process(cmd)
	if err == nil {
		t.Error("Expected error for empty ImageURL, got nil")
	}
}

func TestStubProcessor_Process_AllCommandTypes(t *testing.T) {
	commands := []models.CommandType{
		models.CommandResize,
		models.CommandFilter,
		models.CommandTransform,
		models.CommandAnalyze,
	}

	for _, cmdType := range commands {
		t.Run(string(cmdType), func(t *testing.T) {
			processor := NewStubProcessor()
			cmd := models.ImageCommand{
				ID:       "test-" + string(cmdType),
				Command:  cmdType,
				ImageURL: "https://example.com/image.jpg",
			}

			err := processor.Process(cmd)
			if err != nil {
				t.Errorf("Expected no error for command %s, got: %v", cmdType, err)
			}
		})
	}
}

func TestStubProcessor_GetLastCommand_Empty(t *testing.T) {
	processor := NewStubProcessor()

	lastCmd := processor.GetLastCommand()
	if lastCmd != nil {
		t.Error("Expected nil for empty processor, got command")
	}
}
