package camera

import (
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/lib"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const (
	AutoFocusNormal AutoFocusRange = "normal"
	AutoFocusMacro                 = "macro"
	AutoFocusFull                  = "full"
)

const (
	AutoFocusModeManual AutoFocusMode = "manual"
	AutoFocusModeAuto                 = "auto"
)

const (
	EncodingJPEG   Encoding = "jpg"
	EncodingPNG             = "png"
	EncodingRGB             = "rgb"
	EncodingBMP             = "bmp"
	EncodingYuv420          = "yuv420"
)

const (
	DenoiseAuto    Denoise = "auto"
	DenoiseOff             = "off"
	DenoiseCdnOff          = "cdn_off"
	DenoiseCdnFast         = "cdn_fast"
	DenoiseCdnHq           = "cdn_hq"
)

type AutoFocusRange string

type Encoding string
type Denoise string

type AutoFocusMode string
type CameraSettings struct {
	Width          string
	Height         string
	StreamCodec    string // h264
	AutoFocusRange AutoFocusRange
	AutoFocusMode  AutoFocusMode
	Quality        int
	HDR            bool
	VFlip          bool
	HFlip          bool
	Encoding       Encoding
	Denoise        Denoise
}

type Camera interface {
	TakePhoto(filePath string) error
	Settings() *CameraSettings
	UpdateSettings(settings *CameraSettings)
	OpenStream(port int) (*exec.Cmd, error)
	StopStreaming(streamCmd *exec.Cmd) error
}

type LibCamera struct {
	settings *CameraSettings
}

func (c *LibCamera) Settings() *CameraSettings {
	return c.settings
}

func (c *LibCamera) UpdateSettings(settings *CameraSettings) {
	c.settings = settings
}

func NewLibCamera(settings *CameraSettings) Camera {
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

func getHDR(enabled bool) string {
	if enabled {
		return "auto"
	}
	return "off"
}

func (c *LibCamera) commonArgs() []string {
	return []string{
		"--autofocus-range", string(c.settings.AutoFocusRange),
		"--autofocus-mode", string(c.settings.AutoFocusMode),
		"--vflip", boolToStr(c.settings.VFlip),
		"--hflip", boolToStr(c.settings.HFlip),
		"--hdr", getHDR(c.settings.HDR),
		"-n",
		"-e", string(c.settings.Encoding),
		"-q", strconv.FormatInt(int64(c.settings.Quality), 10),
		"--denoise", string(c.settings.Denoise),
	}
}

func (c *LibCamera) OpenStream(port int) (*exec.Cmd, error) {
	args := append(c.commonArgs(),
		"t", "0",
		"--width", c.settings.Width,
		"--height", c.settings.Height,
		"--codec", c.settings.StreamCodec,
		"--inline", "--listen", "-o", fmt.Sprintf("tcp://0.0.0.0:%d", port),
	)

	theCmd := exec.Command("libcamera-vid", args...)

	theCmd.Stdout = os.Stdout
	theCmd.Stderr = os.Stderr

	return theCmd, theCmd.Start()
}

var ErrNoProcess = fmt.Errorf("No process to kill")

func (s *LibCamera) StopStreaming(streamCmd *exec.Cmd) error {
	if streamCmd != nil && streamCmd.Process != nil {
		err := lib.KillProcess(streamCmd.Process.Pid)
		if err != nil {
			return fmt.Errorf("killing process: %v", err)
		}
		return nil
	}
	return ErrNoProcess
}

func (c *LibCamera) TakePhoto(filePath string) error {
	_ = os.Mkdir(filepath.Dir(filePath), 0755)
	args := append(c.commonArgs(),
		"-o", filePath,
	)

	cmd := exec.Command("libcamera-still", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
