package queue

import (
	"testing"

	pb "kafka-consumer/proto"
)

func TestDecoder_DecodeJSON(t *testing.T) {
	decoder := NewDecoder(FormatJSON)

	jsonData := []byte(`{
		"id": "test-123",
		"command": "resize",
		"image_url": "https://example.com/image.jpg",
		"parameters": {
			"width": 800,
			"height": 600
		}
	}`)

	result, err := decoder.DecodeCommand(jsonData)
	if err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	cmd, ok := result.(*pb.ImageCommand)
	if !ok {
		t.Fatal("Expected *pb.ImageCommand")
	}

	if cmd.Id != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", cmd.Id)
	}

	if cmd.Command != pb.CommandType_COMMAND_TYPE_RESIZE {
		t.Errorf("Expected RESIZE command, got %v", cmd.Command)
	}

	if cmd.ImageUrl != "https://example.com/image.jpg" {
		t.Errorf("Unexpected image URL: %s", cmd.ImageUrl)
	}
}

func TestDecoder_DecodeJSON_AllCommands(t *testing.T) {
	decoder := NewDecoder(FormatJSON)

	testCases := []struct {
		name     string
		command  string
		expected pb.CommandType
	}{
		{"resize", "resize", pb.CommandType_COMMAND_TYPE_RESIZE},
		{"filter", "filter", pb.CommandType_COMMAND_TYPE_FILTER},
		{"transform", "transform", pb.CommandType_COMMAND_TYPE_TRANSFORM},
		{"analyze", "analyze", pb.CommandType_COMMAND_TYPE_ANALYZE},
		{"crop", "crop", pb.CommandType_COMMAND_TYPE_CROP},
		{"remove_background", "remove_background", pb.CommandType_COMMAND_TYPE_REMOVE_BACKGROUND},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData := []byte(`{"id": "test", "command": "` + tc.command + `", "image_url": "https://example.com/img.jpg"}`)
			result, err := decoder.DecodeCommand(jsonData)
			if err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			cmd := result.(*pb.ImageCommand)
			if cmd.Command != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, cmd.Command)
			}
		})
	}
}

func TestDecoder_DecodeInvalidJSON(t *testing.T) {
	decoder := NewDecoder(FormatJSON)

	_, err := decoder.DecodeCommand([]byte("{ invalid json }"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
