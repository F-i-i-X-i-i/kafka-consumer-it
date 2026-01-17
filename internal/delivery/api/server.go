// API1

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"

	"kafka-consumer/internal/models/api"
	"kafka-consumer/internal/pkg/customerrors"
	pb "kafka-consumer/proto"
)

// MessageSender interface for sending messages to Kafka
type MessageSender interface {
	SendMessage(ctx context.Context, cmd *pb.ImageCommand) error
}

// Server represents the HTTP API server for health checks and message sending
type Server struct {
	startTime     time.Time
	messagesCount int64
	producer      MessageSender
	validate      *validator.Validate
	router        chi.Router
}

// NewServer creates a new API server with chi router
func NewServer() *Server {
	s := &Server{
		startTime: time.Now(),
		validate:  validator.New(),
	}
	s.setupRouter()
	return s
}

// setupRouter configures the chi router with all routes
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Routes - method is defined at router level, no need for checks in handlers
	r.Get("/health", s.HealthHandler)
	r.Get("/ready", s.ReadyHandler)
	r.Get("/stats", s.StatsHandler)
	r.Post("/send", s.SendHandler)

	s.router = r
}

// Router returns the chi router for use in http.Server
func (s *Server) Router() http.Handler {
	return s.router
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
	response := api.HealthResponse{
		Status:        "healthy",
		Uptime:        time.Since(s.startTime).String(),
		MessagesCount: s.GetMessagesCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReadyHandler handles readiness check requests
func (s *Server) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.ReadyResponse{Status: "ready"})
}

// StatsHandler returns processing statistics
func (s *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := api.StatsResponse{
		UptimeSeconds:     time.Since(s.startTime).Seconds(),
		MessagesProcessed: s.GetMessagesCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// SendHandler handles requests to send messages to Kafka
func (s *Server) SendHandler(w http.ResponseWriter, r *http.Request) {
	var req api.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		customerrors.WriteError(w, customerrors.ErrBadRequest.WithDetails("Invalid JSON: "+err.Error()))
		return
	}
	log.Printf("📥 API Received: ID=%s, Command=%s, URL=%s, HasParams=%v, Params=%+v",
		req.ID, req.Command, req.ImageURL, 
		req.Parameters != nil, req.Parameters)

	// Validate request using validator
	if err := s.validate.Struct(req); err != nil {
		customerrors.WriteValidationError(w, err.Error())
		return
	}

	// Convert to protobuf command
	cmd := &pb.ImageCommand{
		Id:       req.ID,
		ImageUrl: req.ImageURL,
	}

	// Map command type and parameters
	switch req.Command {
	case "resize":
		cmd.Command = pb.CommandType_COMMAND_TYPE_RESIZE
		if req.Parameters != nil {
			params := &pb.ResizeParameters{}
			if w, ok := req.Parameters["width"].(float64); ok {
				params.Width = int32(w)
			}
			if h, ok := req.Parameters["height"].(float64); ok {
				params.Height = int32(h)
			}
			cmd.Parameters = &pb.ImageCommand_Resize{Resize: params}
		}
	case "filter":
		cmd.Command = pb.CommandType_COMMAND_TYPE_FILTER
		if req.Parameters != nil {
			params := &pb.FilterParameters{}
			if ft, ok := req.Parameters["filter_type"].(string); ok {
				params.FilterType = ft
			}
			if i, ok := req.Parameters["intensity"].(float64); ok {
				params.Intensity = i
			}
			cmd.Parameters = &pb.ImageCommand_Filter{Filter: params}
		}
	case "transform":
		cmd.Command = pb.CommandType_COMMAND_TYPE_TRANSFORM
		if req.Parameters != nil {
			params := &pb.TransformParameters{}
			if rd, ok := req.Parameters["rotation_degrees"].(float64); ok {
				params.RotationDegrees = rd
			}
			if fh, ok := req.Parameters["flip_horizontal"].(bool); ok {
				params.FlipHorizontal = fh
			}
			if fv, ok := req.Parameters["flip_vertical"].(bool); ok {
				params.FlipVertical = fv
			}
			cmd.Parameters = &pb.ImageCommand_Transform{Transform: params}
		}
	case "analyze":
		cmd.Command = pb.CommandType_COMMAND_TYPE_ANALYZE
		if req.Parameters != nil {
			params := &pb.AnalyzeParameters{}
			if models, ok := req.Parameters["models"].([]interface{}); ok {
				for _, m := range models {
					if model, ok := m.(string); ok {
						params.Models = append(params.Models, model)
					}
				}
			}
			cmd.Parameters = &pb.ImageCommand_Analyze{Analyze: params}
		}
	case "crop":
		cmd.Command = pb.CommandType_COMMAND_TYPE_CROP
		if req.Parameters != nil {
			params := &pb.CropParameters{}
			if x, ok := req.Parameters["x"].(float64); ok {
				params.X = int32(x)
			}
			if y, ok := req.Parameters["y"].(float64); ok {
				params.Y = int32(y)
			}
			if w, ok := req.Parameters["width"].(float64); ok {
				params.Width = int32(w)
			}
			if h, ok := req.Parameters["height"].(float64); ok {
				params.Height = int32(h)
			}
			cmd.Parameters = &pb.ImageCommand_Crop{Crop: params}
		}
	case "remove_background":
		cmd.Command = pb.CommandType_COMMAND_TYPE_REMOVE_BACKGROUND
		if req.Parameters != nil {
			params := &pb.RemoveBackgroundParameters{}
			if of, ok := req.Parameters["output_format"].(string); ok {
				params.OutputFormat = of
			}
			if hq, ok := req.Parameters["high_quality"].(bool); ok {
				params.HighQuality = hq
			}
			cmd.Parameters = &pb.ImageCommand_RemoveBackground{RemoveBackground: params}
		}
	default:
		customerrors.WriteError(w, customerrors.ErrInvalidCommand.WithDetails("Unknown command: "+req.Command))
		return
	}

	log.Printf("📤 Sending to Kafka: ID=%s, Command=%v, HasProtoParams=%v",
		cmd.Id, cmd.Command, cmd.Parameters != nil)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.producer.SendMessage(ctx, cmd); err != nil {
		customerrors.WriteError(w, customerrors.ErrInternal.WithDetails("Failed to send message: "+err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(api.SendMessageResponse{
		Success: true,
		Message: "Message sent successfully",
		ID:      req.ID,
	})
	s.IncrementMessagesCount()

}
