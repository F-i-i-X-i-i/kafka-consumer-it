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
		if jsonCmd.Parameters != nil {
			params := &pb.TransformParameters{}
			if rd, ok := jsonCmd.Parameters["rotation_degrees"].(float64); ok {
				params.RotationDegrees = rd
			}
			if fh, ok := jsonCmd.Parameters["flip_horizontal"].(bool); ok {
				params.FlipHorizontal = fh
			}
			if fv, ok := jsonCmd.Parameters["flip_vertical"].(bool); ok {
				params.FlipVertical = fv
			}
			cmd.Parameters = &pb.ImageCommand_Transform{Transform: params}
		}
	case "analyze":
		cmd.Command = pb.CommandType_COMMAND_TYPE_ANALYZE
		if jsonCmd.Parameters != nil {
			params := &pb.AnalyzeParameters{}
			if models, ok := jsonCmd.Parameters["models"].([]interface{}); ok {
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
		if jsonCmd.Parameters != nil {
			params := &pb.CropParameters{}
			if x, ok := jsonCmd.Parameters["x"].(float64); ok {
				params.X = int32(x)
			}
			if y, ok := jsonCmd.Parameters["y"].(float64); ok {
				params.Y = int32(y)
			}
			if w, ok := jsonCmd.Parameters["width"].(float64); ok {
				params.Width = int32(w)
			}
			if h, ok := jsonCmd.Parameters["height"].(float64); ok {
				params.Height = int32(h)
			}
			cmd.Parameters = &pb.ImageCommand_Crop{Crop: params}
		}
	case "remove_background":
		cmd.Command = pb.CommandType_COMMAND_TYPE_REMOVE_BACKGROUND
		if jsonCmd.Parameters != nil {
			params := &pb.RemoveBackgroundParameters{}
			if of, ok := jsonCmd.Parameters["output_format"].(string); ok {
				params.OutputFormat = of
			}
			if hq, ok := jsonCmd.Parameters["high_quality"].(bool); ok {
				params.HighQuality = hq
			}
			cmd.Parameters = &pb.ImageCommand_RemoveBackground{RemoveBackground: params}
		}
	default:
		cmd.Command = pb.CommandType_COMMAND_TYPE_UNSPECIFIED
	}

	return cmd, nil
}
