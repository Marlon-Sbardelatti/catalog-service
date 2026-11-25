package utils

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
)

const maxImageSize = 5 << 20 // 5 MiB

// ReadImage validates the upload and returns the reader. The caller must close it.
func ReadImage(file *multipart.FileHeader) (string, multipart.File, error) {
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", nil, errors.New("image content type required")
	}

	if file.Size > maxImageSize {
		return "", nil, errors.New("image must be under 5MiB")
	}

	src, err := file.Open()
	if err != nil {
		return "", nil, fmt.Errorf("opening image: %w", err)
	}

	return contentType, src, nil
}
