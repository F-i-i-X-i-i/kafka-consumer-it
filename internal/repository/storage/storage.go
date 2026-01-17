package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"bytes"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"kafka-consumer/internal/pkg/logger"
)

// Storage defines the interface for file storage operations
// type Storage interface {
// 	Download(ctx context.Context, key string) (io.ReadCloser, error)
// 	Upload(ctx context.Context, key string, data io.Reader) (string, error)
// 	Delete(ctx context.Context, key string) error
// }

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

// Storage defines the interface for file storage operations
type Storage interface {
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Upload(ctx context.Context, key string, data io.Reader) (string, error)
	Delete(ctx context.Context, key string) error
}

// S3Storage implements Storage interface using S3/MinIO
type S3Storage struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewS3Storage creates a new S3/MinIO storage instance
func NewS3Storage(endpoint, region, accessKey, secretKey, bucket, prefix string) (*S3Storage, error) {
	// Use path-style addressing for MinIO (required)
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for MinIO
	})

	logger.Info("S3/MinIO storage initialized", 
		"endpoint", endpoint, 
		"bucket", bucket,
		"prefix", prefix)

	return &S3Storage{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}, nil
}

// Download downloads a file from S3/MinIO
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullKey := s.getFullKey(key)
	
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", fullKey, err)
	}
	
	logger.Debug("Downloaded file from S3", "key", fullKey)
	return result.Body, nil
}

// Upload uploads a file to S3/MinIO
// Upload uploads a file to S3/MinIO
func (s *S3Storage) Upload(ctx context.Context, key string, data io.Reader) (string, error) {
    fullKey := s.getFullKey(key)
    
    // Читаем все данные в память, чтобы получить размер и сделать seekable reader
    dataBytes, err := io.ReadAll(data)
    if err != nil {
        return "", fmt.Errorf("failed to read data: %w", err)
    }
    
    // Создаем bytes.Reader, который поддерживает Seek
    reader := bytes.NewReader(dataBytes)
    
    _, err = s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:        aws.String(s.bucket),
        Key:           aws.String(fullKey),
        Body:          reader,
        ContentLength: aws.Int64(int64(len(dataBytes))),
        ContentType:   aws.String("image/png"),
    })
    if err != nil {
        return "", fmt.Errorf("failed to put object %s: %w", fullKey, err)
    }

    url := fmt.Sprintf("s3://%s/%s", s.bucket, fullKey)
    logger.Debug("Uploaded file to S3", 
        "key", fullKey, 
        "url", url,
        "size_bytes", len(dataBytes))
    return url, nil
}

// Delete deletes a file from S3/MinIO
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	fullKey := s.getFullKey(key)
	
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", fullKey, err)
	}
	
	logger.Debug("Deleted file from S3", "key", fullKey)
	return nil
}

// getFullKey adds prefix to the key
func (s *S3Storage) getFullKey(key string) string {
	if s.prefix != "" {
		return filepath.Join(s.prefix, key)
	}
	return key
}

// LocalStorage (удалить или оставить как fallback - здесь удаляем)
// type LocalStorage struct{} // УДАЛЯЕМ ЭТУ СТРУКТУРУ

// Удаляем NewLocalStorage и все его методы
// func NewLocalStorage(basePath string) *LocalStorage {} // УДАЛЯЕМ