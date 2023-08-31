package main

import (
	"flag"
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"log"
	"path/filepath"
	"time"
)

const timeFormat = "2006-01-02__15-04-05"

func main() {
	settings := &camera.Settings{
		AutoFocusRange: camera.AutoFocusMacro,
	}

	cam := camera.NewLibCamera(settings)

	var directory string
	flag.StringVar(&directory, "d", "photos", "output directory variable")

	var sleepTime int
	flag.IntVar(&sleepTime, "t", 60, "time to wait after taking photo")

	flag.Parse()

	for {
		fileName := fmt.Sprintf("%s.jpg", time.Now().Format(timeFormat))
		go func() {
			err := cam.TakePhoto(filepath.Join(directory, fileName))
			if err != nil {
				log.Printf("failed to take photo: %v", err)
			}
		}()
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}
