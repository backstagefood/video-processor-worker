package utils

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func GetEnvVarOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func IsValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func SanitizeEmailForPath(email string) string {
	// Replace @ with _
	sanitized := strings.ReplaceAll(email, "@", "_")

	// You might want to add additional sanitization:
	// - Replace dots (.) if they cause issues
	sanitized = strings.ReplaceAll(sanitized, ".", "_")

	// Remove any other potentially problematic characters
	sanitized = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			return r
		}
		return '_' // Replace other special chars with underscore
	}, sanitized)

	return sanitized
}

func ExtractFrames(videoData []byte, fps float64) ([]image.Image, error) {
	// Create a context for timeout control
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the FFmpeg command with proper error handling
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-loglevel", "error", // Only show errors
		"-i", "pipe:0", // Read from stdin
		"-f", "image2pipe", // Output to pipe
		"-vf", fmt.Sprintf("fps=%f", fps), // Frames per second
		"-vcodec", "mjpeg", // Output JPEG frames
		"-q:v", "2", // Quality factor
		"-vsync", "0", // Prevent frame duplication/drop
		"-frame_pts", "1", // Better frame timestamp handling
		"pipe:1", // Write to stdout
	)

	// Set up pipes with buffer
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe error: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe error: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start error: %w", err)
	}

	// Write video data to stdin in a goroutine
	go func() {
		defer stdin.Close()
		if _, err := io.Copy(stdin, bytes.NewReader(videoData)); err != nil {
			log.Printf("stdin write error: %v", err)
		}
	}()

	var frames []image.Image
	scanner := bufio.NewScanner(stdout)
	scanner.Split(scanJPEGFrames) // Custom split function for JPEG frames

	// Scan for JPEG frames
	for scanner.Scan() {
		frameData := scanner.Bytes()
		img, err := jpeg.Decode(bytes.NewReader(frameData))
		if err != nil {
			log.Printf("frame decode error (skipping): %v", err)
			continue
		}
		frames = append(frames, img)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		log.Printf("scanner error: %v", err)
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		// Ignore certain exit codes that FFmpeg may return
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 1 && exitErr.ExitCode() != 183 {
				return frames, fmt.Errorf("ffmpeg exited with code %d: %w",
					exitErr.ExitCode(), err)
			}
			// Consider exit code 1 and 183 as non-fatal
			log.Printf("ffmpeg exited with code %d (non-fatal)", exitErr.ExitCode())
		} else {
			return frames, fmt.Errorf("ffmpeg wait error: %w", err)
		}
	}

	return frames, nil
}

// Custom scanner split function for JPEG frames
func scanJPEGFrames(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find JPEG start marker (0xFFD8)
	if i := bytes.Index(data, []byte{0xFF, 0xD8}); i >= 0 {
		// Find JPEG end marker (0xFFD9) after the start
		if j := bytes.Index(data[i:], []byte{0xFF, 0xD9}); j >= 0 {
			end := i + j + 2 // Include the end marker
			return end, data[i:end], nil
		}
	}

	// Request more data if we haven't found a complete frame
	if !atEOF {
		return 0, nil, nil
	}

	// At EOF but no complete frame found
	return 0, nil, io.ErrUnexpectedEOF
}

// GetBaseFilename extracts just the filename without path or extension
func GetBaseFilename(fullPath string) string {
	// Get the base filename with extension
	filename := filepath.Base(fullPath)

	// Remove the file extension
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]

	return base
}

// GetFileName extrai apenas o nome do arquivo de um caminho completo
func GetFileName(fullPath string) string {
	// Usa filepath.Base para pegar a última parte do caminho
	filename := filepath.Base(fullPath)

	// Remove qualquer barra que possa ter sobrado no final (caso de diretórios)
	filename = strings.TrimSuffix(filename, "/")
	filename = strings.TrimSuffix(filename, "\\")

	return filename
}

func GetFileSize(file multipart.File) (int64, error) {
	// Salva a posição atual do cursor
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	// Move para o final do arquivo para obter o tamanho
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	// Retorna o cursor para a posição original
	_, err = file.Seek(currentPos, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// CreateImageZipInMemory creates a ZIP file in memory containing the provided images
// and returns a multipart.File interface
func CreateImageZipInMemory(images []image.Image) (multipart.File, error) {
	// Create a buffer to hold the ZIP data in memory
	buf := new(bytes.Buffer)

	// Create a new ZIP writer writing to the buffer
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Add each image to the ZIP
	for i, img := range images {
		err := addImageToMemoryZip(zipWriter, img, i)
		if err != nil {
			return nil, err
		}
	}

	// Important: Close the zip writer to flush any pending data
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	// Create a reader from the buffer that implements multipart.File
	file := &bytesFile{
		Reader: bytes.NewReader(buf.Bytes()),
		buf:    buf,
	}

	return file, nil
}

// bytesFile implements multipart.File interface using bytes.Reader
type bytesFile struct {
	*bytes.Reader
	buf *bytes.Buffer
}

func (f *bytesFile) Close() error {
	// No-op for bytes.Buffer
	return nil
}

// We need to implement Seek method which is already provided by bytes.Reader
// The other methods (Read, ReadAt, etc) are also provided by bytes.Reader

// Stat returns file info (dummy implementation)
func (f *bytesFile) Stat() (fileInfo, error) {
	return fileInfo{
		size: int64(f.buf.Len()),
	}, nil
}

// fileInfo implements multipart.FileInfo
type fileInfo struct {
	size int64
}

func (fi fileInfo) Size() int64    { return fi.size }
func (fi fileInfo) IsDir() bool    { return false }
func (fi fileInfo) Mode() uint32   { return 0 }
func (fi fileInfo) ModTime() int64 { return 0 }
func (fi fileInfo) Name() string   { return "images.zip" }
func (fi fileInfo) Sys() any       { return nil }

// addImageToZip remains the same as your original function
func addImageToMemoryZip(zipWriter *zip.Writer, img image.Image, index int) error {
	fileName := "image_" + strconv.Itoa(index) + ".jpg"
	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}
	return jpeg.Encode(fileWriter, img, &jpeg.Options{Quality: 90})
}
