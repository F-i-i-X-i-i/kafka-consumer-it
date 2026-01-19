package usecase

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"time"

	"github.com/disintegration/imaging"

	"kafka-consumer/internal/config"
	"kafka-consumer/internal/pkg/logger"
	"kafka-consumer/internal/pkg/metrics"
	"kafka-consumer/internal/repository/storage"
	pb "kafka-consumer/proto"
)

// ImageProcessor is the interface for image processing operations
type ImageProcessor interface {
	Process(ctx context.Context, cmd *pb.ImageCommand) (*ProcessingResult, error)
}

// ProcessingResult represents the result of image processing
type ProcessingResult struct {
	CommandID        string
	Success          bool
	OutputUrl        string
	ErrorMessage     string
	ProcessingTimeMs int64
}

// RealImageProcessor implements actual image processing with S3 storage
type RealImageProcessor struct {
	storage storage.Storage
}

// NewRealImageProcessor creates a new real image processor with S3 storage
func NewRealImageProcessor(cfg *config.Config) (*RealImageProcessor, error) {
	s3Storage, err := storage.NewS3Storage(
		cfg.S3Endpoint,
		cfg.S3Region,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		cfg.S3Bucket,
		cfg.S3Prefix,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 storage: %w", err)
	}

	logger.Info("S3 storage initialized",
		"endpoint", cfg.S3Endpoint,
		"bucket", cfg.S3Bucket,
		"prefix", cfg.S3Prefix)

	return &RealImageProcessor{
		storage: s3Storage,
	}, nil
}

// Process processes an image command with real operations
func (p *RealImageProcessor) Process(ctx context.Context, cmd *pb.ImageCommand) (*ProcessingResult, error) {
	start := time.Now()
	commandType := cmd.Command.String()

	log := logger.With(
		"command_id", cmd.Id,
		"command_type", commandType,
		"image_url", cmd.ImageUrl,
	)

	log.Info("Processing image command")

	result := &ProcessingResult{
		CommandID: cmd.Id,
	}

	// Validate command
	if cmd.Id == "" {
		metrics.RecordMessageProcessed(commandType, "error")
		result.ErrorMessage = "command ID is required"
		result.Success = false
		return result, fmt.Errorf("command ID is required")
	}
	if cmd.ImageUrl == "" {
		metrics.RecordMessageProcessed(commandType, "error")
		result.ErrorMessage = "image URL is required"
		result.Success = false
		return result, fmt.Errorf("image URL is required")
	}

	// Download image from URL
	img, format, err := p.downloadImage(ctx, cmd.ImageUrl)
	if err != nil {
		log.Error("Failed to download image", "error", err)
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("download failed: %v", err)
		metrics.RecordMessageProcessed(commandType, "error")
		return result, err
	}

	log.Info("Image downloaded", 
		"format", format, 
		"width", img.Bounds().Dx(), 
		"height", img.Bounds().Dy())

	// Process based on command type
	var processedImg image.Image
	switch cmd.Command {
	case pb.CommandType_COMMAND_TYPE_RESIZE:
		processedImg, err = p.processResize(img, cmd.GetResize())
	case pb.CommandType_COMMAND_TYPE_FILTER:
		processedImg, err = p.processFilter(img, cmd.GetFilter())
	case pb.CommandType_COMMAND_TYPE_TRANSFORM:
		processedImg, err = p.processTransform(img, cmd.GetTransform())
	case pb.CommandType_COMMAND_TYPE_CROP:
		processedImg, err = p.processCrop(img, cmd.GetCrop())
	case pb.CommandType_COMMAND_TYPE_ANALYZE:
		// For analyze, we just return image info without saving
		result.Success = true
		result.OutputUrl = ""
		result.ProcessingTimeMs = time.Since(start).Milliseconds()
		log.Info("Image analyzed", 
			"width", img.Bounds().Dx(), 
			"height", img.Bounds().Dy())
		metrics.RecordMessageProcessed(commandType, "success")
		return result, nil
	case pb.CommandType_COMMAND_TYPE_REMOVE_BACKGROUND:
		// Placeholder for AI background removal
		log.Warn("Remove background not implemented, applying grayscale instead")
		processedImg = imaging.Grayscale(img)
	default:
		err = fmt.Errorf("unknown command type: %s", cmd.Command)
	}

	if err != nil {
		log.Error("Processing failed", "error", err)
		result.Success = false
		result.ErrorMessage = err.Error()
		metrics.RecordMessageProcessed(commandType, "error")
		return result, err
	}

	// Encode processed image to PNG bytes
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, processedImg, imaging.PNG); err != nil {
		log.Error("Failed to encode processed image", "error", err)
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("encode failed: %v", err)
		metrics.RecordMessageProcessed(commandType, "error")
		return result, err
	}

	// Upload to S3 storage
	s3Key := fmt.Sprintf("%s.png", cmd.Id)
	outputUrl, err := p.storage.Upload(ctx, s3Key, bytes.NewReader(buf.Bytes()))
	if err != nil {
		log.Error("Failed to upload to S3", "error", err)
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("upload failed: %v", err)
		metrics.RecordMessageProcessed(commandType, "error")
		return result, err
	}

	duration := time.Since(start)
	result.Success = true
	result.OutputUrl = outputUrl
	result.ProcessingTimeMs = duration.Milliseconds()

	metrics.ObserveMessageProcessingDuration(commandType, duration.Seconds())
	metrics.RecordMessageProcessed(commandType, "success")

	log.Info("Image processed successfully",
		"output_url", outputUrl,
		"duration_ms", result.ProcessingTimeMs,
		"image_size", buf.Len(),
		"s3_key", s3Key)

	return result, nil
}

// downloadImage downloads an image from URL
func (p *RealImageProcessor) downloadImage(ctx context.Context, url string) (image.Image, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	// Set user-agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Kafka Image Processor)")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Limit read to 50MB
	limitedReader := io.LimitReader(resp.Body, 50*1024*1024)

	// Decode image
	img, format, err := image.Decode(limitedReader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	return img, format, nil
}

// processResize resizes the image
func (p *RealImageProcessor) processResize(img image.Image, params *pb.ResizeParameters) (image.Image, error) {
	if params == nil {
		return nil, fmt.Errorf("resize parameters required")
	}

	width := int(params.Width)
	height := int(params.Height)

	if width <= 0 && height <= 0 {
		return nil, fmt.Errorf("at least one of width or height must be positive")
	}

	var resized image.Image
	if params.MaintainAspectRatio || width == 0 || height == 0 {
		// Maintain aspect ratio
		if width == 0 {
			resized = imaging.Resize(img, 0, height, imaging.Lanczos)
		} else if height == 0 {
			resized = imaging.Resize(img, width, 0, imaging.Lanczos)
		} else {
			resized = imaging.Fit(img, width, height, imaging.Lanczos)
		}
	} else {
		// Exact resize
		resized = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	return resized, nil
}

// processFilter applies filters to the image
func (p *RealImageProcessor) processFilter(img image.Image, params *pb.FilterParameters) (image.Image, error) {
	if params == nil {
		return nil, fmt.Errorf("filter parameters required")
	}

	intensity := params.Intensity
	if intensity <= 0 {
		intensity = 1.0
	}

	switch params.FilterType {
	case "blur":
		sigma := 3.0 * intensity
		return imaging.Blur(img, sigma), nil
	case "sharpen":
		sigma := 3.0 * intensity
		return imaging.Sharpen(img, sigma), nil
	case "grayscale":
		return imaging.Grayscale(img), nil
	case "invert":
		return imaging.Invert(img), nil
	case "brightness":
		// intensity from -100 to 100
		adjustment := intensity*200 - 100 // Convert 0.0-1.0 to -100 to 100
		return imaging.AdjustBrightness(img, adjustment), nil
	case "contrast":
		// intensity from -100 to 100
		adjustment := intensity*200 - 100
		return imaging.AdjustContrast(img, adjustment), nil
	case "saturation":
		// intensity from -100 to 100
		adjustment := intensity*200 - 100
		return imaging.AdjustSaturation(img, adjustment), nil
	default:
		return nil, fmt.Errorf("unknown filter type: %s", params.FilterType)
	}
}

// processTransform applies transformations to the image
func (p *RealImageProcessor) processTransform(img image.Image, params *pb.TransformParameters) (image.Image, error) {
	if params == nil {
		return img, nil
	}

	result := img

	// Apply rotation
	if params.RotationDegrees != 0 {
		result = imaging.Rotate(result, params.RotationDegrees, image.Transparent)
	}

	// Apply flips
	if params.FlipHorizontal {
		result = imaging.FlipH(result)
	}
	if params.FlipVertical {
		result = imaging.FlipV(result)
	}

	return result, nil
}

// processCrop crops the image
func (p *RealImageProcessor) processCrop(img image.Image, params *pb.CropParameters) (image.Image, error) {
	if params == nil {
		return nil, fmt.Errorf("crop parameters required")
	}

	x := int(params.X)
	y := int(params.Y)
	width := int(params.Width)
	height := int(params.Height)

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("crop width and height must be positive")
	}

	bounds := img.Bounds()
	if x < 0 || y < 0 || x+width > bounds.Dx() || y+height > bounds.Dy() {
		return nil, fmt.Errorf("crop region out of bounds")
	}

	rect := image.Rect(x, y, x+width, y+height)
	return imaging.Crop(img, rect), nil
}