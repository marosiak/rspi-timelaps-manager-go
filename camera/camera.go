package camera

import (
	"fmt"
	"os"
	"os/exec"
)

type AutoFocusRange string

const (
	AutoFocusNormal AutoFocusRange = "normal"
	AutoFocusMacro                 = "macro"
	AutoFocusFull                  = "full"
)

type Settings struct {
	AutoFocusRange AutoFocusRange
}

type Camera interface {
	TakePhoto(filePath string) error
}

type LibCamera struct {
	settings *Settings
}

func NewLibCamera(settings *Settings) Camera {
	return &LibCamera{
		settings: settings,
	}
}

func (c *LibCamera) TakePhoto(filePath string) error {
	cmd := exec.Command("libcamera-still",
		"-o", filePath,
		"--autofocus-range", "macro",
		"--vflip",
		"--hflip",
		"--hdr", "0",
		"-q", "100",
	)

	// Przekieruj wyjście błędów i standardowe do konsoli, jeśli chcesz je zobaczyć
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to take photo: %w", err)
	}

	return nil
}
