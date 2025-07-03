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
	"log/slog"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func GetEnvVarOrDefault[T any](key string, defaultValue T) T {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	var result T
	var err error

	switch any(defaultValue).(type) {
	case string:
		result = any(valueStr).(T)
	case int:
		var val int
		val, err = strconv.Atoi(valueStr)
		result = any(val).(T)
	case int64:
		var val int64
		val, err = strconv.ParseInt(valueStr, 10, 64)
		result = any(val).(T)
	case float64:
		var val float64
		val, err = strconv.ParseFloat(valueStr, 64)
		result = any(val).(T)
	case bool:
		var val bool
		val, err = strconv.ParseBool(valueStr)
		result = any(val).(T)
	default:
		// For unsupported types, return the default value
		return defaultValue
	}

	if err != nil {
		return defaultValue
	}

	return result
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
	if len(videoData) == 0 {
		return nil, fmt.Errorf("videoData está vazio")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-loglevel", "error",
		"-i", "pipe:0",
		"-f", "image2pipe",
		"-vf", fmt.Sprintf("fps=%f", fps),
		"-vcodec", "mjpeg",
		"-q:v", "2",
		"-vsync", "0",
		"-frame_pts", "1",
		"pipe:1",
	)

	// Configuração de pipes (igual ao seu código original)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe error: %w", err)
	}

	// Inicia o FFmpeg
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start error: %w", err)
	}

	// Escreve os dados no stdin em uma goroutine
	go func() {
		defer stdin.Close()
		if _, err := io.Copy(stdin, bytes.NewReader(videoData)); err != nil {
			slog.Error("falha ao escrever no stdin", "error", err)
		}
	}()

	// Decodifica os frames (igual ao seu código original)
	var frames []image.Image
	scanner := bufio.NewScanner(stdout)
	scanner.Split(scanJPEGFrames)

	for scanner.Scan() {
		img, err := jpeg.Decode(bytes.NewReader(scanner.Bytes()))
		if err != nil {
			slog.Error("frame corrompido (pulando)", "error", err)
			continue
		}
		frames = append(frames, img)
	}

	// Verifica erros do scanner
	if err := scanner.Err(); err != nil {
		slog.Error("erro no scanner", "error", err)
	}

	// Aguarda o término do FFmpeg e trata erros de forma mais robusta
	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 1:
				// Código 1 pode ser crítico (ex.: vídeo inválido)
				if len(frames) == 0 {
					return nil, fmt.Errorf("ffmpeg falhou (código 1): vídeo inválido ou sem frames")
				}
				slog.Warn("ffmpeg concluiu com avisos (código 1), mas alguns frames foram extraídos")
			case 183:
				// Código 183 pode ocorrer em vídeos com problemas de timestamp
				slog.Warn("ffmpeg detectou problemas de timestamp (código 183)")
			default:
				return frames, fmt.Errorf("ffmpeg falhou com código %d: %w", exitErr.ExitCode(), err)
			}
		} else {
			return frames, fmt.Errorf("erro ao aguardar ffmpeg: %w", err)
		}
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("nenhum frame foi extraído (vídeo inválido ou vazio?)")
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
