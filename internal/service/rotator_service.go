package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bielrocafndz/LogShredder/internal/domain"
)

// rotator is the private concrete implementation of domain.RotatorService
type rotator struct{}

// NewRotator creates a new instance of the rotator service
func NewRotator() domain.RotatorService {
	return &rotator{}
}

// Check verifies the file status and decides if rotation is needed
func (r *rotator) Check(path string, max_size int64) (*domain.LogFile, error) {
	file_info, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", domain.ErrLogNotFound, path)
		}
		return nil, fmt.Errorf("system error accessing log: %v", err)
	}

	log_file := &domain.LogFile{
		Path: path,
		Size: file_info.Size(),
	}

	// Trigger domain error if size is exceeded
	if log_file.Size > max_size {
		return log_file, domain.ErrFileTooLarge
	}

	return log_file, nil
}

// spinner displays an animated character in the terminal
func spinner(delay time.Duration, stop_chan chan bool) {
	frames := []string{"\\", "|", "/", "-"}
	for {
		for _, frame := range frames {
			select {
			case <-stop_chan:
				return
			default:
				fmt.Printf("\r[%s] Compressing log file...", frame)
				time.Sleep(delay)
			}
		}
	}
}

// Rotate handles the file compression logic with a visual spinner
func (r *rotator) Rotate(log_file *domain.LogFile) error {
	stop_spinner := make(chan bool)

	// Start the spinner in a background goroutine
	go spinner(100*time.Millisecond, stop_spinner)

	// Execute ZIP logic
	err := compress_to_zip(log_file.Path)

	// Stop the spinner and clear the line
	stop_spinner <- true
	fmt.Print("\r\033[K") // \033[K clears the entire line

	if err != nil {
		return fmt.Errorf("rotation failed: %w", err)
	}

	fmt.Printf("✅ Success: %s rotated and compressed.\n", log_file.Path)
	return nil
}

// compress_to_zip is a private helper to handle the archive/zip logic
func compress_to_zip(source string) error {
	destination := source + ".zip"
	zip_file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer zip_file.Close()

	archive := zip.NewWriter(zip_file)
	defer archive.Close()

	file_to_zip, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file_to_zip.Close()

	info, _ := file_to_zip.Stat()
	header, _ := zip.FileInfoHeader(info)
	header.Name = filepath.Base(source)
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file_to_zip)
	if err != nil {
		return err
	}

	// Explicitly close before removing source to avoid Windows file locks
	archive.Close()
	file_to_zip.Close()

	return os.Remove(source)
}