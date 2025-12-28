package usecase

import (
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"

	"kafka-consumer/internal/pkg/logger"
	"kafka-consumer/internal/pkg/metrics"
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
	OutputPath       string
	ErrorMessage     string
	ProcessingTimeMs int64
}

// RealImageProcessor implements actual image processing
type RealImageProcessor struct {
	outputDir string
}

// NewRealImageProcessor creates a new real image processor
func NewRealImageProcessor(outputDir string) *RealImageProcessor {
	// Create output directory if it doesn't exist
	os.MkdirAll(outputDir, 0755)
	return &RealImageProcessor{
		outputDir: outputDir,
	}
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
		return nil, fmt.Errorf("command ID is required")
	}
	if cmd.ImageUrl == "" {
		metrics.RecordMessageProcessed(commandType, "error")
		return nil, fmt.Errorf("image URL is required")
	}

	// Download image
	img, format, err := p.downloadImage(ctx, cmd.ImageUrl)
	if err != nil {
		log.Error("Failed to download image", "error", err)
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("download failed: %v", err)
		metrics.RecordMessageProcessed(commandType, "error")
		return result, err
	}

	log.Info("Image downloaded", "format", format, "width", img.Bounds().Dx(), "height", img.Bounds().Dy())

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
		// For analyze, we just return image info
		result.Success = true
		result.OutputPath = ""
		result.ProcessingTimeMs = time.Since(start).Milliseconds()
		log.Info("Image analyzed", "width", img.Bounds().Dx(), "height", img.Bounds().Dy())
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

	// Save processed image
	outputPath := filepath.Join(p.outputDir, fmt.Sprintf("%s_processed.png", cmd.Id))
	if err := imaging.Save(processedImg, outputPath); err != nil {
		log.Error("Failed to save processed image", "error", err)
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("save failed: %v", err)
		metrics.RecordMessageProcessed(commandType, "error")
		return result, err
	}

	duration := time.Since(start)
	result.Success = true
	result.OutputPath = outputPath
	result.ProcessingTimeMs = duration.Milliseconds()

	metrics.ObserveMessageProcessingDuration(commandType, duration.Seconds())
	metrics.RecordMessageProcessed(commandType, "success")

	log.Info("Image processed successfully",
		"output_path", outputPath,
		"duration_ms", result.ProcessingTimeMs)

	return result, nil
}

// downloadImage downloads an image from URL
func (p *RealImageProcessor) downloadImage(ctx context.Context, url string) (image.Image, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

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
		return imaging.AdjustBrightness(img, intensity*100-50), nil
	case "contrast":
		// intensity from -100 to 100
		return imaging.AdjustContrast(img, intensity*100-50), nil
	case "saturation":
		// intensity from -100 to 100
		return imaging.AdjustSaturation(img, intensity*100-50), nil
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
