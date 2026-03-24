package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/bielrocafndz/LogShredder/internal/domain"
	"github.com/bielrocafndz/LogShredder/internal/service"
	"github.com/briandowns/spinner"
)

func main() {
	service.LogSync("--- LogShredder initialized ---")

	shredder := service.NewRotator()
	log_path := "app.log"
	max_size := int64(1024 * 5) // 5KB for testing

	// Setup professional spinner (Charset 9 is a clean rotating line)
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Searching for changes..."

	// Function to encapsulate the scanning logic
	scan := func() {
		log_file, err := shredder.Check(log_path, max_size)

		if err != nil {
			if errors.Is(err, domain.ErrFileTooLarge) {
				s.Stop() // Pause UI to print clean log

				timestamp := time.Now().Format("15:04:05")
				service.LogSync(fmt.Sprintf("[WARN] %s | Size limit exceeded (%d bytes)", timestamp, log_file.Size))

				s.Suffix = " Rotating and compressing..."
				s.Start()

				if err_rotate := shredder.Rotate(log_file); err_rotate == nil {
					s.Stop()
					service.LogSync(fmt.Sprintf("[DONE] %s | Processed: %s -> %s.zip", timestamp, log_path, log_path))
				} else {
					s.Stop()
					service.LogSync(fmt.Sprintf("[FAIL] %s | Error: %v", timestamp, err_rotate))
				}

				s.Suffix = " Searching for changes..."
				s.Start()
			}
			// If ErrLogNotFound, we just stay silent and keep the spinner moving
			return
		}
	}

	// Initial execution (no waiting 30s)
	scan()
	s.Start()

	// Periodic execution every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		scan()
	}
}
