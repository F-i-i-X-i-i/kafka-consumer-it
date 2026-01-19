package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pb "kafka-consumer/proto"
)

// mockProducer implements MessageSender for testing
type mockProducer struct {
	lastCommand *pb.ImageCommand
	shouldError bool
}

func (m *mockProducer) SendMessage(ctx context.Context, cmd *pb.ImageCommand) error {
	m.lastCommand = cmd
	if m.shouldError {
		return context.DeadlineExceeded
	}
	return nil
}

func TestServer_HealthHandler(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&response)

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}
}

func TestServer_ReadyHandler(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	server.ReadyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestServer_StatsHandler(t *testing.T) {
	server := NewServer()
	server.IncrementMessagesCount()
	server.IncrementMessagesCount()

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()

	server.StatsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&response)

	if response["messages_processed"] != float64(2) {
		t.Errorf("Expected 2 messages, got %v", response["messages_processed"])
	}
}

func TestServer_SendHandler_Success(t *testing.T) {
	server := NewServer()
	producer := &mockProducer{}
	server.SetProducer(producer)

	body := `{"id": "test-001", "command": "resize", "image_url": "https://example.com/test.jpg"}`
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.SendHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	if producer.lastCommand == nil {
		t.Error("Expected producer to receive command")
	}

	if producer.lastCommand.Id != "test-001" {
		t.Errorf("Expected ID 'test-001', got '%s'", producer.lastCommand.Id)
	}
}

func TestServer_SendHandler_ValidationError(t *testing.T) {
	server := NewServer()
	producer := &mockProducer{}
	server.SetProducer(producer)

	// Missing required fields
	body := `{"id": "", "command": "resize"}`
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.SendHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestServer_SendHandler_InvalidJSON(t *testing.T) {
	server := NewServer()
	producer := &mockProducer{}
	server.SetProducer(producer)

	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader("{ invalid }"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.SendHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestServer_Router(t *testing.T) {
	server := NewServer()
	router := server.Router()

	if router == nil {
		t.Error("Expected router to not be nil")
	}
}

func TestServer_IncrementMessagesCount(t *testing.T) {
	server := NewServer()

	if server.GetMessagesCount() != 0 {
		t.Error("Expected initial count to be 0")
	}

	server.IncrementMessagesCount()
	server.IncrementMessagesCount()
	server.IncrementMessagesCount()

	if server.GetMessagesCount() != 3 {
		t.Errorf("Expected count 3, got %d", server.GetMessagesCount())
	}
}
