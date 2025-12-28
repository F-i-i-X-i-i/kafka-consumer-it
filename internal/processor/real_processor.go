package processor

import (
	"context"
	"fmt"

	"kafka-consumer/internal/usecase"
	pb "kafka-consumer/proto"
)

// RealProcessor wraps usecase.ImageProcessor to implement Processor interface
type RealProcessor struct {
	imageProcessor usecase.ImageProcessor
}

// NewRealProcessor creates a new real processor wrapping the image processor
func NewRealProcessor(outputDir string) *RealProcessor {
	return &RealProcessor{
		imageProcessor: usecase.NewRealImageProcessor(outputDir),
	}
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
