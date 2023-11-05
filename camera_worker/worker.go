package camera_worker

import (
	"errors"
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/api"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/rs/zerolog/log"
	"os/exec"
	"path/filepath"
	"time"
)

type CameraWorker struct {
	camera    camera.Camera
	cfg       *config.Config
	streamCmd *exec.Cmd
	pubSub    *api.PubSub
}

func NewCameraWorker(camera camera.Camera, cfg *config.Config, pubSub *api.PubSub) *CameraWorker {
	return &CameraWorker{camera: camera, cfg: cfg, pubSub: pubSub}
}

const timeFormat = "2006-01-02__15-04-05"

func (w *CameraWorker) configToCameraSettings() {
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
}

func (w *CameraWorker) takePhoto() {
	w.configToCameraSettings()

	if w.cfg.Streaming {
		w.stopStreaming()
	}

	fileName := fmt.Sprintf("%s.%s", time.Now().Format(timeFormat), w.cfg.Encoding)
	err := w.camera.TakePhoto(filepath.Join(w.cfg.OutputDir, fileName))
	if err != nil {
		log.Printf("failed to take photo: %v", err)
	} else {
		err := w.pubSub.PublishJson(api.PhotosTopic, api.PhotoResponse{
			Photo:     fileName,
			CreatedAt: time.Now().Unix(),
		})
		if err != nil {
			log.Err(err).Msg("notify subscribers about new photo")
		}
	}

	if w.cfg.Streaming {
		w.openStream()
	}
}

func (w *CameraWorker) stopStreaming() {
	err := w.camera.StopStreaming(w.streamCmd)
	if err != nil && !errors.Is(err, camera.ErrNoProcess) {
		log.Printf("failed to stop streamCmd: %v", err)
	} else {
		w.streamCmd = nil
	}
}

func (w *CameraWorker) openStream() {
	var err error
	w.configToCameraSettings()
	if w.streamCmd == nil {
		log.Debug().Msg("Opening camera stream")
		go func() {
			w.streamCmd, err = w.camera.OpenStream(8888)
			if err != nil {
				log.Printf("failed to open streamCmd: %v", err)
			}
		}()
	}
}

func (w *CameraWorker) Run() {
	if w.cfg.Streaming {
		w.openStream()
	}
	for {
		go w.takePhoto()
		time.Sleep(w.cfg.Delay)
	}
}
