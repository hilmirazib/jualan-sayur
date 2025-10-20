package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"user-service/internal/core/port"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SupabaseStorage struct {
	projectURL    string
	apiKey        string
	bucketName    string
	httpClient    *http.Client
}

type SupabaseUploadResponse struct {
	Key string `json:"Key"`
}

func NewSupabaseStorage(projectURL, apiKey, bucketName string) (port.StorageInterface, error) {
	if projectURL == "" || apiKey == "" || bucketName == "" {
		return nil, fmt.Errorf("supabase project URL, API key, and bucket name are required")
	}

	return &SupabaseStorage{
		projectURL:    strings.TrimSuffix(projectURL, "/"),
		apiKey:        apiKey,
		bucketName:    bucketName,
		httpClient:    &http.Client{},
	}, nil
}

func (s *SupabaseStorage) UploadFile(ctx context.Context, bucketName, objectName string, file io.Reader, contentType string) (string, error) {
	// If bucketName is empty, use the default bucket
	if bucketName == "" {
		bucketName = s.bucketName
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

	// Read file content
	log.Info().Str("content_type", contentType).Msg("[SupabaseStorage-UploadFile] Starting to read file content")
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("[SupabaseStorage-UploadFile] Failed to read file content")
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	log.Info().Int("content_length", len(fileContent)).Msg("[SupabaseStorage-UploadFile] File content read successfully")

	// Check if file content is empty
	if len(fileContent) == 0 {
		log.Error().Msg("[SupabaseStorage-UploadFile] File content is empty after reading")
		return "", fmt.Errorf("file content is empty")
	}

	// Create multipart form data
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add file field
	fw, err := w.CreateFormFile("file", objectName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = fw.Write(fileContent); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	// Close the writer
	w.Close()

	// Create upload URL
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, bucketName, objectName)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &b)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Cache-Control", "max-age=31536000") // 1 year cache

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Generate public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.projectURL, bucketName, objectName)

	return publicURL, nil
}

func (s *SupabaseStorage) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	// If bucketName is empty, use the default bucket
	if bucketName == "" {
		bucketName = s.bucketName
	}

	// Create delete URL
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, bucketName, objectName)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
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
