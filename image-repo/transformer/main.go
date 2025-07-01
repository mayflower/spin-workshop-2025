package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"

	"github.com/anthonynsimon/bild/transform"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

type ErrorResponse struct {
	Reason string `json:"reason,omitempty"`
}

func parsePositiveInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	if val <= 0 {
		return 0, fmt.Errorf("value must be positive")
	}

	return val, nil
}

func sendResponseError(w http.ResponseWriter, reason string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := ErrorResponse{Reason: reason}
	json.NewEncoder(w).Encode(response)
}

func sendResponsePng(w http.ResponseWriter, image image.Image) {
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	png.Encode(w, image)
}

func sendResponseJpeg(w http.ResponseWriter, image image.Image) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	jpeg.Encode(w, image, nil)
}

func sendImageResponse(w http.ResponseWriter, image image.Image, mimeType string) {
	switch mimeType {
	case "image/jpeg":
		sendResponseJpeg(w, image)
	default:
		sendResponsePng(w, image)
	}
}

func readImage(bodyReader io.ReadCloser) (image.Image, error) {
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read request body: %v", err)
	}

	reader := bytes.NewReader(body)
	reader.Seek(0, io.SeekStart)

	_, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, fmt.Errorf("Unable to determine image format. Only JPEG and PNG are supported.")
	}

	reader.Seek(0, io.SeekStart)

	var img image.Image
	switch format {
	case "jpeg":
		img, err = jpeg.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse JPEG image: %v", err)
		}
	case "png":
		img, err = png.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse PNG image: %v", err)
		}
	default:
		return nil, fmt.Errorf("Unsupported image format: %s. Only JPEG and PNG are supported.", format)
	}

	return img, nil
}

func handleTransform(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendResponseError(w, "Only POST method is allowed", http.StatusBadRequest)
		return
	}

	img, err := readImage(r.Body)
	if err != nil {
		sendResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	bounds := img.Bounds()
	origWidth := bounds.Max.X - bounds.Min.X
	origHeight := bounds.Max.Y - bounds.Min.Y

	widthStr := r.URL.Query().Get("width")
	heightStr := r.URL.Query().Get("height")
	typeStr := r.URL.Query().Get("type")

	if typeStr == "" {
		typeStr = "image/png"
	} else if typeStr != "image/png" && typeStr != "image/jpeg" {
		sendResponseError(w, fmt.Sprintf("Unsupported image type: %s. Only image/png and image/jpeg are supported.", typeStr), http.StatusBadRequest)
		return
	}

	if widthStr == "" && heightStr == "" {
		sendImageResponse(w, img, typeStr)
		return
	}

	width, err := parsePositiveInt(widthStr)
	if err != nil {
		sendResponseError(w, fmt.Sprintf("Invalid width parameter: %v", err), http.StatusBadRequest)
		return
	}

	height, err := parsePositiveInt(heightStr)
	if err != nil {
		sendResponseError(w, fmt.Sprintf("Invalid height parameter: %v", err), http.StatusBadRequest)
		return
	}

	if width == 0 && height > 0 {
		width = int(float64(origWidth) * float64(height) / float64(origHeight))
	} else if height == 0 && width > 0 {
		height = int(float64(origHeight) * float64(width) / float64(origWidth))
	}

	resizedImg := transform.Resize(img, width, height, transform.Linear)

	sendImageResponse(w, resizedImg, typeStr)
}

func init() {
	spinhttp.Handle(handleTransform)
}
