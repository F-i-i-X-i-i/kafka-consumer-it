package models

// CommandType represents the type of image processing command
type CommandType string

const (
	CommandResize    CommandType = "resize"
	CommandFilter    CommandType = "filter"
	CommandTransform CommandType = "transform"
	CommandAnalyze   CommandType = "analyze"
)

// ImageCommand represents a command to process an image with AI
type ImageCommand struct {
	ID         string                 `json:"id"`
	Command    CommandType            `json:"command"`
	ImageURL   string                 `json:"image_url"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ResizeParameters represents parameters for resize command
type ResizeParameters struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// FilterParameters represents parameters for filter command
type FilterParameters struct {
	FilterType string  `json:"filter_type"`
	Intensity  float64 `json:"intensity"`
}
