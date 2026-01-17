package processor

import (
	"context"
	"fmt"

	"kafka-consumer/internal/config"
	"kafka-consumer/internal/usecase"
	pb "kafka-consumer/proto"
)

// RealProcessor wraps usecase.ImageProcessor to implement Processor interface
type RealProcessor struct {
	imageProcessor usecase.ImageProcessor
}

// NewRealProcessor creates a new real processor wrapping the image processor
func NewRealProcessor(cfg *config.Config) (*RealProcessor, error) {
	imageProcessor, err := usecase.NewRealImageProcessor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create image processor: %w", err)
	}
	
	return &RealProcessor{
		imageProcessor: imageProcessor,
	}, nil
}

// ProcessAny handles interface{} type from decoder
func (p *RealProcessor) ProcessAny(ctx context.Context, cmd interface{}) error {
	pbCmd, ok := cmd.(*pb.ImageCommand)
	if !ok {
		return fmt.Errorf("invalid command type: expected *pb.ImageCommand")
	}
	return p.Process(ctx, pbCmd)
}

// Process processes the command using the real image processor
func (p *RealProcessor) Process(ctx context.Context, cmd *pb.ImageCommand) error {
	_, err := p.imageProcessor.Process(ctx, cmd)
	return err
}
