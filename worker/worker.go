package worker

import (
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/rs/zerolog/log"
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
	w.cfg = config.New()
	w.camera.UpdateSettings(&camera.Settings{
		AutoFocusRange: w.cfg.AutoFocusRange,
		Quality:        w.cfg.Quality,
		HDR:            w.cfg.Hdr,
		VFlip:          w.cfg.VFlip,
		HFlip:          w.cfg.HFlip,
		Encoding:       w.cfg.Encoding,
		Denoise:        w.cfg.Denoise,
	})
	log.Debug().Msg("Updated camera settings")
}

func (w *Worker) Record() {
	for {
		fileName := fmt.Sprintf("%s.%s", time.Now().Format(timeFormat), w.cfg.Encoding)
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
