// API1

package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"

	"kafka-consumer/internal/pkg/logger"
	pb "kafka-consumer/proto"
)

// MessageFormat represents the format of Kafka messages
type MessageFormat string

const (
	FormatJSON     MessageFormat = "json"
	FormatProtobuf MessageFormat = "protobuf"
)

// Producer represents a Kafka producer for sending image commands
type Producer struct {
	writer *kafka.Writer
	format MessageFormat
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
		format: FormatJSON, // Default to JSON for backward compatibility
	}
}

// SetMessageFormat sets the message format for encoding
func (p *Producer) SetMessageFormat(format MessageFormat) {
	p.format = format
}

// SendMessage sends an image command to Kafka
func (p *Producer) SendMessage(ctx context.Context, cmd *pb.ImageCommand) error {
	var data []byte
	var err error

	if p.format == FormatProtobuf {
		data, err = proto.Marshal(cmd)
	} else {
		data, err = p.encodeJSON(cmd)
	}

	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(cmd.Id),
		Value: data,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		logger.Error("Error sending message", "error", err)
		return err
	}

	logger.Info("Message sent", "command_id", cmd.Id, "command_type", cmd.Command.String())
	return nil
}

// encodeJSON encodes protobuf command to JSON
func (p *Producer) encodeJSON(cmd *pb.ImageCommand) ([]byte, error) {
	jsonCmd := map[string]interface{}{
		"id":        cmd.Id,
		"command":   commandTypeToString(cmd.Command),
		"image_url": cmd.ImageUrl,
	}

	// Add parameters based on type
	params := make(map[string]interface{})
	switch p := cmd.Parameters.(type) {
	case *pb.ImageCommand_Resize:
		params["width"] = p.Resize.Width
		params["height"] = p.Resize.Height
	case *pb.ImageCommand_Filter:
		params["filter_type"] = p.Filter.FilterType
		params["intensity"] = p.Filter.Intensity
	case *pb.ImageCommand_Transform:
		params["rotation_degrees"] = p.Transform.RotationDegrees
	case *pb.ImageCommand_Crop:
		params["x"] = p.Crop.X
		params["y"] = p.Crop.Y
		params["width"] = p.Crop.Width
		params["height"] = p.Crop.Height
	}

	if len(params) > 0 {
		jsonCmd["parameters"] = params
	}

	return json.Marshal(jsonCmd)
}

func commandTypeToString(ct pb.CommandType) string {
	switch ct {
	case pb.CommandType_COMMAND_TYPE_RESIZE:
		return "resize"
	case pb.CommandType_COMMAND_TYPE_FILTER:
		return "filter"
	case pb.CommandType_COMMAND_TYPE_TRANSFORM:
		return "transform"
	case pb.CommandType_COMMAND_TYPE_ANALYZE:
		return "analyze"
	case pb.CommandType_COMMAND_TYPE_CROP:
		return "crop"
	case pb.CommandType_COMMAND_TYPE_REMOVE_BACKGROUND:
		return "remove_background"
	default:
		return "unspecified"
	}
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
