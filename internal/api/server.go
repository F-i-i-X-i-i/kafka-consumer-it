package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"kafka-consumer/internal/models"
)

// MessageSender interface for sending messages to Kafka
type MessageSender interface {
	SendMessage(ctx context.Context, cmd models.ImageCommand) error
}

// Server represents the HTTP API server for health checks and message sending
type Server struct {
	startTime      time.Time
	messagesCount  int64
	kafkaConnected bool
	producer       MessageSender
}

// NewServer creates a new API server
func NewServer() *Server {
	return &Server{
		startTime:      time.Now(),
		kafkaConnected: false,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status         string `json:"status"`
	Uptime         string `json:"uptime"`
	KafkaConnected bool   `json:"kafka_connected"`
	MessagesCount  int64  `json:"messages_processed"`
}

// SendMessageRequest represents a request to send a message to Kafka
type SendMessageRequest struct {
	ID         string                 `json:"id"`
	Command    string                 `json:"command"`
	ImageURL   string                 `json:"image_url"`
	Parameters map[string]interface{} `json:"parameters"`
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// SetKafkaConnected updates the Kafka connection status
func (s *Server) SetKafkaConnected(connected bool) {
	s.kafkaConnected = connected
}

// SetProducer sets the Kafka producer for sending messages
func (s *Server) SetProducer(producer MessageSender) {
	s.producer = producer
}

// IncrementMessagesCount increments the processed messages counter
func (s *Server) IncrementMessagesCount() {
	atomic.AddInt64(&s.messagesCount, 1)
}

// GetMessagesCount returns the current messages count
func (s *Server) GetMessagesCount() int64 {
	return atomic.LoadInt64(&s.messagesCount)
}

// HealthHandler handles health check requests
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := "healthy"
	if !s.kafkaConnected {
		status = "degraded"
	}

	response := HealthResponse{
		Status:         status,
		Uptime:         time.Since(s.startTime).String(),
		KafkaConnected: s.kafkaConnected,
		MessagesCount:  s.GetMessagesCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReadyHandler handles readiness check requests
func (s *Server) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.kafkaConnected {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"reason": "kafka_disconnected",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

// StatsHandler returns processing statistics
func (s *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"uptime_seconds":     time.Since(s.startTime).Seconds(),
		"messages_processed": s.GetMessagesCount(),
		"kafka_connected":    s.kafkaConnected,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// SendHandler handles requests to send messages to Kafka
func (s *Server) SendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.producer == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Message: "Producer not configured",
		})
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// Validate request
	if req.ID == "" || req.Command == "" || req.ImageURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Message: "Missing required fields: id, command, image_url",
		})
		return
	}

	cmd := models.ImageCommand{
		ID:         req.ID,
		Command:    models.CommandType(req.Command),
		ImageURL:   req.ImageURL,
		Parameters: req.Parameters,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.producer.SendMessage(ctx, cmd); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Message: "Failed to send message: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SendMessageResponse{
		Success: true,
		Message: "Message sent successfully",
		ID:      req.ID,
	})
}
