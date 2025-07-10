package utils

import (
	"bytes"
	"image"
	"image/color"
	"mime/multipart"
	"os"
	"testing"
)

func TestGetEnvVarOrDefault(t *testing.T) {
	os.Setenv("TEST_KEY", "123")
	defer os.Unsetenv("TEST_KEY")

	// Test with existing environment variable
	result := GetEnvVarOrDefault("TEST_KEY", 0)
	if result != 123 {
		t.Errorf("Expected 123, got %v", result)
	}

	// Test with non-existing environment variable
	result = GetEnvVarOrDefault("NON_EXISTENT_KEY", 42)
	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestIsValidVideoFile(t *testing.T) {
	validFiles := []string{"video.mp4", "movie.avi", "clip.mov"}
	invalidFiles := []string{"image.jpg", "document.pdf", "audio.mp3"}

	for _, file := range validFiles {
		if !IsValidVideoFile(file) {
			t.Errorf("Expected %s to be valid", file)
		}
	}

	for _, file := range invalidFiles {
		if IsValidVideoFile(file) {
			t.Errorf("Expected %s to be invalid", file)
		}
	}
}

func TestSanitizeEmailForPath(t *testing.T) {
	email := "user.name@example.com"
	expected := "user_name_example_com"
	result := SanitizeEmailForPath(email)

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetBaseFilename(t *testing.T) {
	path := "/path/to/file.txt"
	expected := "file"
	result := GetBaseFilename(path)

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetFileName(t *testing.T) {
	path := "/path/to/file.txt"
	expected := "file.txt"
	result := GetFileName(path)

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetFileSize(t *testing.T) {
	content := []byte("test content")
	file := multipart.File(&bytesFile{Reader: bytes.NewReader(content), buf: bytes.NewBuffer(content)})

	size, err := GetFileSize(file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), size)
	}
}

func TestCreateImageZipInMemory(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	images := []image.Image{img}
	file, err := CreateImageZipInMemory(images)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	size, err := GetFileSize(file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if size == 0 {
		t.Errorf("Expected non-zero file size, got %d", size)
	}
}
