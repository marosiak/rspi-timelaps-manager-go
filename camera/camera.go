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

const (
	DenoiseAuto    Denoise = "auto"
	DenoiseOff             = "off"
	DenoiseCdnOff          = "cdn_off"
	DenoiseCdnFast         = "cdn_fast"
	DenoiseCdnHq           = "cdn_hq"
)
type Denoise string

type AutoFocusMode string
type Settings struct {
	AutoFocusRange AutoFocusRange
	AutoFocusMode AutoFocusMode
	Quality        int
	HDR            bool
	VFlip          bool
	HFlip          bool
	Encoding       Encoding
	Denoise        Denoise
}

type Camera interface {
	TakePhoto(filePath string) error
	Settings() *Settings
	UpdateSettings(settings *Settings)
}

type LibCamera struct {
	settings *Settings
}

func (c *LibCamera) Settings() *Settings {
	return c.settings
}

func (c *LibCamera) UpdateSettings(settings *Settings) {
	c.settings = settings
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
		"--autofocus-mode", string(c.settings.AutoFocusMode)
		"--vflip", boolToStr(c.settings.VFlip),
		"--hflip", boolToStr(c.settings.HFlip),
		"--hdr", boolToStr(c.settings.HDR),
		"-n",
		"-e", string(c.settings.Encoding),
		"-q", strconv.FormatInt(int64(c.settings.Quality), 10),
		"--denoise", string(c.settings.Denoise),
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
