package worker

import (
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"log"
	"path/filepath"
	"time"
)

type Worker struct {
	camera camera.Camera
	cfg    *config.Config
}

func NewWorker(camera camera.Camera, cfg *config.Config) *Worker {
	return &Worker{camera: camera, cfg: cfg}
}

const timeFormat = "2006-01-02__15-04-05"

func (w *Worker) configToCameraSettings() {
	w.camera.Settings().AutoFocusRange = w.cfg.AutoFocusRange
	w.camera.Settings().Quality = w.cfg.Quality
	w.camera.Settings().HDR = w.cfg.HDR
	w.camera.Settings().VFlip = w.cfg.VFlip
	w.camera.Settings().HFlip = w.cfg.HFlip
	w.camera.Settings().Encoding = w.cfg.Encoding
}

func (w *Worker) Record() {
	for {
		fileName := fmt.Sprintf("%s.jpg", time.Now().Format(timeFormat))
		go func() {
			w.configToCameraSettings()
			err := w.camera.TakePhoto(filepath.Join(w.cfg.OutputDir, fileName))
			if err != nil {
				log.Printf("failed to take photo: %v", err)
			}
		}()
		time.Sleep(w.cfg.Delay)
	}
}
