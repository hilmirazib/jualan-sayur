package port

import (
	"context"
	"io"
)

type StorageInterface interface {
	UploadFile(ctx context.Context, bucketName, objectName string, file io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
}
