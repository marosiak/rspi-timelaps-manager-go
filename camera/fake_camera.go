package camera

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FakeCamera struct {
	settings *Settings
}

func (c *FakeCamera) Settings() *Settings {
	return c.settings
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

func NewFakeCamera() Camera {
	return &FakeCamera{}
}
