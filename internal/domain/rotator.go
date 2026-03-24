package domain

import "errors"

// === EXCEPTIONS (Domain Errors) ===
var (
	ErrLogNotFound		= errors.New("Log file does not exist on the specified path")
	ErrFileTooLarge		= errors.New("File is over allowed size")
	ErrInvalidPolicy	= errors.New("Configured retention policy is not valid")
)

// === MODELS ===
// LogFile: entity we work with
type LogFile struct {
	path string
	size int64
}

// === INTERFACES ===
// RotatorService defines which actions our rotator can perform.
type RotatorService interface  {
	check(path string, max_size int64) (*LogFile, error)
	Rotate(file *LogFile) error
}