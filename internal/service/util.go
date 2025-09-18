package service

import (
	"fmt"
	"mime"
	"mime/multipart"
	"path"
	"path/filepath"
	"strings"
	"time"
	"upload-util/internal/config"

	"github.com/google/uuid"
)

func generateFileName(header *multipart.FileHeader, settings *config.UploadSettings) string {
	originalName := header.Filename
	extension := filepath.Ext(originalName)
	nameWithoutExtension := strings.TrimSuffix(originalName, extension)

	var newName string
	switch settings.FilenameStrategy {
	case "uuid":
		newName = uuid.New().String()
	case "timestamp":
		newName = fmt.Sprintf("%d", time.Now().Unix())
	case "original":
		newName = nameWithoutExtension
	default:
		newName = uuid.New().String()

	}
	if settings.KeepOriginalName && settings.FilenameStrategy != "original" {
		newName = fmt.Sprintf("%s_%s", nameWithoutExtension, newName)
	}
	return newName + extension
}

func validateFile(header *multipart.FileHeader, settings *config.UploadSettings) error {
	maxSize := settings.MaxFileSize * 1024 * 1024
	if header.Size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", header.Size, maxSize)
	}
	if len(settings.AllowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := false
		for _, allowedExt := range settings.AllowedExtensions {
			if strings.ToLower(allowedExt) == ext {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension %s is not allowed", ext)
		}
	}
	return nil
}

func getMimeType(filename string) string {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

func buildObjectKey(filename, pathPrefix string) string {
	if pathPrefix == "" {
		return filename
	}
	return path.Join(pathPrefix, filename)
}

func buildURL(domain, bucket, endpoint, key string, useSSL bool) string {
	if domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(domain, "/"), key)
	}

	protocol := "http"
	if useSSL {
		protocol = "https"
	}

	if strings.Contains(endpoint, bucket) {
		return fmt.Sprintf("%s://%s/%s", protocol, endpoint, key)
	}
	return fmt.Sprintf("%s://%s.%s/%s", protocol, bucket, endpoint, key)
}
