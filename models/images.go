package models

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ImageSourceType defines the type of image source
type ImageSourceType string

const (
	// Base64ImageSource represents a base64-encoded image source
	Base64ImageSource ImageSourceType = "base64"

	// URLImageSource represents a URL-based image source
	URLImageSource ImageSourceType = "url"
)

// MediaType defines image media types
type MediaType string

const (
	// JPEGMediaType represents JPEG images
	JPEGMediaType MediaType = "image/jpeg"

	// PNGMediaType represents PNG images
	PNGMediaType MediaType = "image/png"

	// GIFMediaType represents GIF images
	GIFMediaType MediaType = "image/gif"

	// WebPMediaType represents WebP images
	WebPMediaType MediaType = "image/webp"
)

// ImageSource represents the source of an image
type ImageSource struct {
	Type      ImageSourceType `json:"type"`
	MediaType MediaType       `json:"media_type,omitempty"`
	Data      string          `json:"data,omitempty"`
	URL       string          `json:"url,omitempty"`
}

// NewBase64ImageSource creates a new base64-encoded image source
func NewBase64ImageSource(mediaType MediaType, data string) ImageSource {
	return ImageSource{
		Type:      Base64ImageSource,
		MediaType: mediaType,
		Data:      data,
	}
}

// NewURLImageSource creates a new URL-based image source
func NewURLImageSource(url string) ImageSource {
	return ImageSource{
		Type: URLImageSource,
		URL:  url,
	}
}

// CreateImageBlock creates a new image content block
func CreateImageBlock(source ImageSource) ContentBlock {
	return ContentBlock{
		ImageContent: &ImageBlock{
			Type:   ImageContentType,
			Source: source,
		},
	}
}

// Base64EncodeImage encodes an image file as base64
func Base64EncodeImage(filePath string) (string, MediaType, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("error reading file: %w", err)
	}

	mediaType := http.DetectContentType(data)
	switch mediaType {
	case string(JPEGMediaType), string(PNGMediaType), string(GIFMediaType), string(WebPMediaType):
	default:
		return "", "", fmt.Errorf("unsupported media type: %s", mediaType)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return encoded, MediaType(mediaType), nil
}
