package api

// HealthResponse represents the health check response
type HealthResponse struct {
	Status        string `json:"status"`
	Uptime        string `json:"uptime"`
	MessagesCount int64  `json:"messages_processed"`
}

// ReadyResponse represents the readiness check response
type ReadyResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// StatsResponse represents the stats response
type StatsResponse struct {
	UptimeSeconds     float64 `json:"uptime_seconds"`
	MessagesProcessed int64   `json:"messages_processed"`
}

// SendMessageRequest represents a request to send a message to Kafka
type SendMessageRequest struct {
	ID         string                 `json:"id" validate:"required"`
	Command    string                 `json:"command" validate:"required,oneof=resize filter transform analyze crop remove_background"`
	ImageURL   string                 `json:"image_url" validate:"required,url"`
	Parameters map[string]interface{} `json:"parameters"`
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}
