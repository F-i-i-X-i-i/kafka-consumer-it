package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"kafka-consumer/internal/pkg/logger"
)

// Storage defines the interface for file storage operations
type Storage interface {
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Upload(ctx context.Context, key string, data io.Reader) (string, error)
	Delete(ctx context.Context, key string) error
}

// LocalStorage implements Storage interface using local filesystem
// This is a mock/development implementation
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath string) *LocalStorage {
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath}
}

// Download downloads a file from local storage
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	path := filepath.Join(s.basePath, key)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	logger.Debug("Downloaded file from local storage", "key", key, "path", path)
	return file, nil
}

// Upload uploads a file to local storage
func (s *LocalStorage) Upload(ctx context.Context, key string, data io.Reader) (string, error) {
	path := filepath.Join(s.basePath, key)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", path, err)
	}

	logger.Debug("Uploaded file to local storage", "key", key, "path", path)
	return path, nil
}

// Delete deletes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	path := filepath.Join(s.basePath, key)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	logger.Debug("Deleted file from local storage", "key", key, "path", path)
	return nil
}

// S3Storage implements Storage interface using AWS S3
// This is a placeholder for real S3 implementation
type S3Storage struct {
	bucket   string
	region   string
	endpoint string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(bucket, region, endpoint string) *S3Storage {
	return &S3Storage{
		bucket:   bucket,
		region:   region,
		endpoint: endpoint,
	}
}

// Download downloads a file from S3
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// TODO: Implement real S3 download using aws-sdk-go-v2
	// For now, return an error indicating not implemented
	logger.Warn("S3 download not implemented, use LOCAL_STORAGE=true for development")
	return nil, fmt.Errorf("S3 storage not implemented: bucket=%s, key=%s", s.bucket, key)
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(ctx context.Context, key string, data io.Reader) (string, error) {
	// TODO: Implement real S3 upload using aws-sdk-go-v2
	logger.Warn("S3 upload not implemented, use LOCAL_STORAGE=true for development")
	return "", fmt.Errorf("S3 storage not implemented: bucket=%s, key=%s", s.bucket, key)
}

// Delete deletes a file from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	// TODO: Implement real S3 delete using aws-sdk-go-v2
	logger.Warn("S3 delete not implemented, use LOCAL_STORAGE=true for development")
	return fmt.Errorf("S3 storage not implemented: bucket=%s, key=%s", s.bucket, key)
}
