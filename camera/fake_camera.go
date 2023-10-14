package camera

import (
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/lib"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type FakeCamera struct {
	settings *CameraSettings
}

func (c *FakeCamera) Settings() *CameraSettings {
	if c.settings == nil {
		c.settings = &CameraSettings{}
	}

	return c.settings
}

func (c *FakeCamera) OpenStream(port int) (*exec.Cmd, error) {
	theCmd := exec.Command("gst-launch-1.0",
		"videotestsrc", "!",
		"x264enc", "!",
		"tcpserversink", "host=0.0.0.0", fmt.Sprintf("port=%d", port),
	)

	theCmd.Stdout = os.Stdout
	theCmd.Stderr = os.Stderr

	return theCmd, theCmd.Start()
}

func (c *FakeCamera) TakePhoto(filePath string) error {
	_ = os.Mkdir(filepath.Dir(filePath), 0755)

	resp, err := http.Get("https://placekitten.com/256/256")
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *FakeCamera) StopStreaming(streamCmd *exec.Cmd) error {
	if streamCmd != nil && streamCmd.Process != nil {
		err := lib.KillProcess(streamCmd.Process.Pid)
		if err != nil {
			return fmt.Errorf("killing process: %v", err)
		}
		return nil
	}
	return ErrNoProcess
}

func (c *FakeCamera) UpdateSettings(settings *CameraSettings) {}

func NewFakeCamera() Camera {
	return &FakeCamera{}
}
