package queue

import (
	"encoding/json"
	"fmt"

	pb "kafka-consumer/proto"

	"google.golang.org/protobuf/proto"
)

// MessageFormat represents the format of Kafka messages
type MessageFormat string

const (
	FormatJSON     MessageFormat = "json"
	FormatProtobuf MessageFormat = "protobuf"
)

// Decoder implements MessageDecoder interface
type Decoder struct {
	preferredFormat MessageFormat
}

// NewDecoder creates a new message decoder
func NewDecoder(format MessageFormat) *Decoder {
	return &Decoder{
		preferredFormat: format,
	}
}

// DecodeCommand decodes a message payload into an ImageCommand
func (d *Decoder) DecodeCommand(data []byte) (interface{}, error) {
	if d.preferredFormat == FormatProtobuf {
		cmd := &pb.ImageCommand{}
		if err := proto.Unmarshal(data, cmd); err == nil {
			return cmd, nil
		}
		// Fall back to JSON
		return d.decodeJSON(data)
	}

	// Try JSON first
	cmd, err := d.decodeJSON(data)
	if err == nil {
		return cmd, nil
	}

	// Fall back to Protobuf
	pbCmd := &pb.ImageCommand{}
	if err := proto.Unmarshal(data, pbCmd); err != nil {
		return nil, fmt.Errorf("failed to decode message: neither JSON nor Protobuf")
	}
	return pbCmd, nil
}

// decodeJSON decodes JSON message to protobuf command
func (d *Decoder) decodeJSON(data []byte) (*pb.ImageCommand, error) {
	var jsonCmd struct {
		ID         string                 `json:"id"`
		Command    string                 `json:"command"`
		ImageURL   string                 `json:"image_url"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.Unmarshal(data, &jsonCmd); err != nil {
		return nil, err
	}

	cmd := &pb.ImageCommand{
		Id:       jsonCmd.ID,
		ImageUrl: jsonCmd.ImageURL,
	}

	// Map command type
	switch jsonCmd.Command {
	case "resize":
		cmd.Command = pb.CommandType_COMMAND_TYPE_RESIZE
		if jsonCmd.Parameters != nil {
			params := &pb.ResizeParameters{}
			if w, ok := jsonCmd.Parameters["width"].(float64); ok {
				params.Width = int32(w)
			}
			if h, ok := jsonCmd.Parameters["height"].(float64); ok {
				params.Height = int32(h)
			}
			cmd.Parameters = &pb.ImageCommand_Resize{Resize: params}
		}
	case "filter":
		cmd.Command = pb.CommandType_COMMAND_TYPE_FILTER
		if jsonCmd.Parameters != nil {
			params := &pb.FilterParameters{}
			if ft, ok := jsonCmd.Parameters["filter_type"].(string); ok {
				params.FilterType = ft
			}
			if i, ok := jsonCmd.Parameters["intensity"].(float64); ok {
				params.Intensity = i
			}
			cmd.Parameters = &pb.ImageCommand_Filter{Filter: params}
		}
	case "transform":
		cmd.Command = pb.CommandType_COMMAND_TYPE_TRANSFORM
	case "analyze":
		cmd.Command = pb.CommandType_COMMAND_TYPE_ANALYZE
	case "crop":
		cmd.Command = pb.CommandType_COMMAND_TYPE_CROP
	case "remove_background":
		cmd.Command = pb.CommandType_COMMAND_TYPE_REMOVE_BACKGROUND
	default:
		cmd.Command = pb.CommandType_COMMAND_TYPE_UNSPECIFIED
	}

	return cmd, nil
}
