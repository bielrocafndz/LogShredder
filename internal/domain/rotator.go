package domain

import "errors"

// Custom errors for domain logic
var (
	ErrLogNotFound   = errors.New("log file does not exist in the specified path")
	ErrFileTooLarge  = errors.New("file exceeds the maximum allowed size")
	ErrInvalidPolicy = errors.New("the configured retention policy is invalid")
)

// log_file represents our core entity
type LogFile struct {
	Path string
	Size int64
}

// RotatorService defines the contract for our log rotator
type RotatorService interface {
	Check(path string, max_size int64) (*LogFile, error)
	Rotate(log_file *LogFile) error
}
