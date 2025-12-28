package processor

import (
	"context"
	"testing"

	pb "kafka-consumer/proto"
)

func TestStubProcessor_Process_Success(t *testing.T) {
	processor := NewStubProcessor()
	ctx := context.Background()

	cmd := &pb.ImageCommand{
		Id:       "test-123",
		Command:  pb.CommandType_COMMAND_TYPE_RESIZE,
		ImageUrl: "https://example.com/image.jpg",
		Parameters: &pb.ImageCommand_Resize{
			Resize: &pb.ResizeParameters{
				Width:  100,
				Height: 100,
			},
		},
	}

	err := processor.Process(ctx, cmd)
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
	if lastCmd.Id != cmd.Id {
		t.Errorf("Expected ID %s, got: %s", cmd.Id, lastCmd.Id)
	}
}

func TestStubProcessor_Process_EmptyID(t *testing.T) {
	processor := NewStubProcessor()
	ctx := context.Background()

	cmd := &pb.ImageCommand{
		Id:       "",
		Command:  pb.CommandType_COMMAND_TYPE_RESIZE,
		ImageUrl: "https://example.com/image.jpg",
	}

	err := processor.Process(ctx, cmd)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestStubProcessor_Process_EmptyImageURL(t *testing.T) {
	processor := NewStubProcessor()
	ctx := context.Background()

	cmd := &pb.ImageCommand{
		Id:       "test-123",
		Command:  pb.CommandType_COMMAND_TYPE_RESIZE,
		ImageUrl: "",
	}

	err := processor.Process(ctx, cmd)
	if err == nil {
		t.Error("Expected error for empty ImageURL, got nil")
	}
}

func TestStubProcessor_Process_AllCommandTypes(t *testing.T) {
	commands := []pb.CommandType{
		pb.CommandType_COMMAND_TYPE_RESIZE,
		pb.CommandType_COMMAND_TYPE_FILTER,
		pb.CommandType_COMMAND_TYPE_TRANSFORM,
		pb.CommandType_COMMAND_TYPE_ANALYZE,
		pb.CommandType_COMMAND_TYPE_CROP,
		pb.CommandType_COMMAND_TYPE_REMOVE_BACKGROUND,
	}

	for _, cmdType := range commands {
		t.Run(cmdType.String(), func(t *testing.T) {
			processor := NewStubProcessor()
			ctx := context.Background()
			cmd := &pb.ImageCommand{
				Id:       "test-" + cmdType.String(),
				Command:  cmdType,
				ImageUrl: "https://example.com/image.jpg",
			}

			err := processor.Process(ctx, cmd)
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
