package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"user-service/internal/core/port"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type GCSStorage struct {
	client     *storage.Client
	projectID  string
	bucketName string
}

func NewGCSStorage(projectID, bucketName, credentialsFile string) (port.StorageInterface, error) {
	ctx := context.Background()

	var client *storage.Client
	var err error

	if credentialsFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	} else {
		// Use default credentials (for GCP environments)
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorage{
		client:     client,
		projectID:  projectID,
		bucketName: bucketName,
	}, nil
}

func (g *GCSStorage) UploadFile(ctx context.Context, bucketName, objectName string, file io.Reader, contentType string) (string, error) {
	// If bucketName is empty, use the default bucket
	if bucketName == "" {
		bucketName = g.bucketName
	}

	// Generate unique filename if not provided
	if objectName == "" {
		ext := ".jpg" // default extension
		if contentType != "" {
			switch contentType {
			case "image/jpeg":
				ext = ".jpg"
			case "image/png":
				ext = ".png"
			case "image/gif":
				ext = ".gif"
			case "image/webp":
				ext = ".webp"
			}
		}
		objectName = fmt.Sprintf("profile-%s%s", uuid.New().String(), ext)
	}

	// Create object handle
	obj := g.client.Bucket(bucketName).Object(objectName)

	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=31536000" // 1 year cache

	// Upload file
	if _, err := io.Copy(writer, file); err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Make the object publicly readable
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set ACL: %w", err)
	}

	// Generate public URL
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)

	return publicURL, nil
}

func (g *GCSStorage) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	// If bucketName is empty, use the default bucket
	if bucketName == "" {
		bucketName = g.bucketName
	}

	// Create object handle
	obj := g.client.Bucket(bucketName).Object(objectName)

	// Delete object
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Helper function to validate image file
func ValidateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// Check file size (max 5MB)
	const maxSize = 5 << 20 // 5MB
	if header.Size > maxSize {
		return fmt.Errorf("file size too large, maximum 5MB")
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"}

	validType := false
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid file type, only JPEG, PNG, GIF, and WebP are allowed")
	}

	// Additional validation: check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}

	if !validExt {
		return fmt.Errorf("invalid file extension")
	}

	return nil
}
