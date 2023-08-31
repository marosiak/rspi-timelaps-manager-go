package camera

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type AutoFocusRange string

const (
	AutoFocusNormal AutoFocusRange = "normal"
	AutoFocusMacro                 = "macro"
	AutoFocusFull                  = "full"
)

const (
	EncodingJPEG   Encoding = "jpg"
	EncodingPNG             = "png"
	EncodingRGB             = "rgb"
	EncodingBMP             = "bmp"
	EncodingYuv420          = "yuv420"
)

type Encoding string

type Settings struct {
	AutoFocusRange AutoFocusRange
	Quality        int
	HDR            bool
	VFlip          bool
	HFlip          bool
	Encoding       Encoding
}

type Camera interface {
	TakePhoto(filePath string) error
	Settings() *Settings
}

type LibCamera struct {
	settings *Settings
}

func (c *LibCamera) Settings() *Settings {
	return c.settings
}

func NewLibCamera(settings *Settings) Camera {
	return &LibCamera{
		settings: settings,
	}
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (c *LibCamera) TakePhoto(filePath string) error {
	_ = os.Mkdir(filepath.Dir(filePath), 0755)

	cmd := exec.Command("libcamera-still",
		"-o", filePath,
		"--autofocus-range", string(c.settings.AutoFocusRange),
		"--vflip", boolToStr(c.settings.VFlip),
		"--hflip", boolToStr(c.settings.HFlip),
		"--hdr", boolToStr(c.settings.HDR),
		"-e", string(c.settings.Encoding),
		"-q", strconv.FormatInt(int64(c.settings.Quality), 10),
	)
	println(cmd.String())
	// Przekieruj wyjście błędów i standardowe do konsoli, jeśli chcesz je zobaczyć
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
