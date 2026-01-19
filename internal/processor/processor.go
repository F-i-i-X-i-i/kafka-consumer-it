package processor

import (
	"context"
	"fmt"
	"time"

	"kafka-consumer/internal/pkg/logger"
	"kafka-consumer/internal/pkg/metrics"
	"kafka-consumer/internal/pkg/tracing"
	pb "kafka-consumer/proto"
)

// Processor defines the interface for processing image commands
type Processor interface {
	Process(ctx context.Context, cmd *pb.ImageCommand) error
	// ProcessAny accepts interface{} for use with generic decoders
	ProcessAny(ctx context.Context, cmd interface{}) error
}

// StubProcessor is a stub implementation that logs commands for demonstration
type StubProcessor struct {
	ProcessedCommands []*pb.ImageCommand
}

// NewStubProcessor creates a new stub processor
func NewStubProcessor() *StubProcessor {
	return &StubProcessor{
		ProcessedCommands: make([]*pb.ImageCommand, 0),
	}
}

// ProcessAny handles interface{} type from decoder
func (p *StubProcessor) ProcessAny(ctx context.Context, cmd interface{}) error {
	pbCmd, ok := cmd.(*pb.ImageCommand)
	if !ok {
		return fmt.Errorf("invalid command type: expected *pb.ImageCommand")
	}
	return p.Process(ctx, pbCmd)
}

// Process logs the command and stores it for later inspection
func (p *StubProcessor) Process(ctx context.Context, cmd *pb.ImageCommand) error {
	// Start tracing span
	ctx, span := tracing.StartSpan(ctx, "processor.Process")
	defer span.End()

	start := time.Now()
	commandType := cmd.Command.String()

	log := logger.With(
		"command_id", cmd.Id,
		"command_type", commandType,
		"image_url", cmd.ImageUrl,
		"trace_id", tracing.TraceID(ctx),
	)

	log.Info("Processing command (stub)")

	// Validate command
	if cmd.Id == "" {
		err := fmt.Errorf("command ID is required")
		tracing.RecordError(ctx, err)
		metrics.RecordMessageProcessed(commandType, "error")
		return err
	}
	if cmd.ImageUrl == "" {
		err := fmt.Errorf("image URL is required")
		tracing.RecordError(ctx, err)
		metrics.RecordMessageProcessed(commandType, "error")
		return err
	}

	// Store command for inspection (useful for testing)
	p.ProcessedCommands = append(p.ProcessedCommands, cmd)

	// Log processing based on command type
	switch cmd.Command {
	case pb.CommandType_COMMAND_TYPE_RESIZE:
		if resize := cmd.GetResize(); resize != nil {
			log.Info("Would resize image",
				"width", resize.Width,
				"height", resize.Height)
		}
	case pb.CommandType_COMMAND_TYPE_FILTER:
		if filter := cmd.GetFilter(); filter != nil {
			log.Info("Would apply filter",
				"filter_type", filter.FilterType,
				"intensity", filter.Intensity)
		}
	case pb.CommandType_COMMAND_TYPE_TRANSFORM:
		if transform := cmd.GetTransform(); transform != nil {
			log.Info("Would transform image",
				"rotation", transform.RotationDegrees)
		}
	case pb.CommandType_COMMAND_TYPE_CROP:
		if crop := cmd.GetCrop(); crop != nil {
			log.Info("Would crop image",
				"width", crop.Width,
				"height", crop.Height)
		}
	default:
		log.Info("Command type", "type", commandType)
	}

	duration := time.Since(start).Seconds()
	metrics.ObserveMessageProcessingDuration(commandType, duration)
	metrics.RecordMessageProcessed(commandType, "success")

	log.Info("Command processed successfully", "duration_seconds", duration)
	return nil
}

// GetProcessedCount returns the number of processed commands
func (p *StubProcessor) GetProcessedCount() int {
	return len(p.ProcessedCommands)
}

// GetLastCommand returns the last processed command
func (p *StubProcessor) GetLastCommand() *pb.ImageCommand {
	if len(p.ProcessedCommands) == 0 {
		return nil
	}
	return p.ProcessedCommands[len(p.ProcessedCommands)-1]
}
