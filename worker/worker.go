package worker

import (
	"errors"
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/rs/zerolog/log"
	"os/exec"
	"path/filepath"
	"time"
)

type Worker struct {
	camera    camera.Camera
	cfg       *config.Config
	streamCmd *exec.Cmd
}

func NewWorker(camera camera.Camera, cfg *config.Config) *Worker {
	return &Worker{camera: camera, cfg: cfg}
}

const timeFormat = "2006-01-02__15-04-05"

func (w *Worker) configToCameraSettings() {
	w.cfg = config.New()
	w.camera.UpdateSettings(&camera.CameraSettings{
		AutoFocusRange: w.cfg.AutoFocusRange,
		AutoFocusMode:  w.cfg.AutoFocusMode,

		Quality:  w.cfg.Quality,
		HDR:      w.cfg.Hdr,
		VFlip:    w.cfg.VFlip,
		HFlip:    w.cfg.HFlip,
		Encoding: w.cfg.Encoding,
		Denoise:  w.cfg.Denoise,
	})
	log.Debug().Msg("Updated camera settings")
}

func (w *Worker) takePhoto() {
	w.configToCameraSettings()

	if w.cfg.Streaming {
		w.stopStreaming()
	}

	fileName := fmt.Sprintf("%s.%s", time.Now().Format(timeFormat), w.cfg.Encoding)
	err := w.camera.TakePhoto(filepath.Join(w.cfg.OutputDir, fileName))
	if err != nil {
		log.Printf("failed to take photo: %v", err)
	}

	if w.cfg.Streaming {
		w.openStream()
	}
}

func (w *Worker) stopStreaming() {
	err := w.camera.StopStreaming(w.streamCmd)
	if err != nil && !errors.Is(err, camera.ErrNoProcess) {
		log.Printf("failed to stop streamCmd: %v", err)
	} else {
		w.streamCmd = nil
	}
}

func (w *Worker) openStream() {
	var err error
	w.configToCameraSettings()
	if w.streamCmd == nil {
		log.Debug().Msg("Opening fake camera stream")
		go func() {
			w.streamCmd, err = w.camera.OpenStream(8888)
			if err != nil {
				log.Printf("failed to open streamCmd: %v", err)
			}
		}()
	}
}

func (w *Worker) Run() {
	w.openStream()
	for {
		go w.takePhoto()
		time.Sleep(w.cfg.Delay)
	}
}
