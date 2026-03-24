package service

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/bielrocafndz/LogShredder/internal/domain"
)

type rotator struct{}

func NewRotator() domain.RotatorService {
	return &rotator{}
}

// LogSync ensures the message goes to both stdout and shredder.log
func LogSync(message string) {
	// Print to console
	fmt.Println(message)

	// Write to file
	f, err := os.OpenFile("shredder.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	logger := log.New(f, "", 0)
	logger.Println(message)
}

func (r *rotator) Check(path string, max_size int64) (*domain.LogFile, error) {
	file_info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrLogNotFound
		}
		return nil, err
	}

	log_file := &domain.LogFile{
		Path: path,
		Size: file_info.Size(),
	}

	if log_file.Size > max_size {
		return log_file, domain.ErrFileTooLarge
	}

	return log_file, nil
}

func (r *rotator) Rotate(log_file *domain.LogFile) error {
	// Generamos el nombre del ZIP (Por ahora simple, luego añadiremos Timestamp)
	dest := log_file.Path + ".zip"

	err := compress(log_file.Path, dest)
	if err != nil {
		return err
	}

	return os.Remove(log_file.Path)
}

func compress(source, target string) error {
	zip_file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zip_file.Close()

	archive := zip.NewWriter(zip_file)
	defer archive.Close()

	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	info, _ := file.Stat()
	header, _ := zip.FileInfoHeader(info)
	header.Name = filepath.Base(source)
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
