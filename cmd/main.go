package main

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/macrosiak/rspi-timelaps-manager-go/worker"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func getLatestFile(dir string) string {
	files, _ := os.ReadDir(dir)
	var newestFile string
	var newestTime int64 = 0
	for _, f := range files {
		fi, err := os.Stat(dir + f.Name())
		if err != nil {
			fmt.Println(err)
		}
		currTime := fi.ModTime().Unix()
		if currTime > newestTime {
			newestTime = currTime
			newestFile = f.Name()
		}
	}
	return newestFile
}

func main() {
	cfg := config.New()
	var cam camera.Camera
	if cfg.Development {
		cam = camera.NewFakeCamera()
	} else {
		settings := &camera.Settings{
			AutoFocusRange: camera.AutoFocusMacro,
		}
		cam = camera.NewLibCamera(settings)
	}

	var directory string
	flag.StringVar(&directory, "d", "photos", "output directory variable")

	var sleepTime int
	flag.IntVar(&sleepTime, "t", 60, "time to wait after taking photo")

	flag.Parse()

	cfg.Delay = time.Duration(sleepTime) * time.Second
	cfg.OutputDir = directory

	timelapseWorker := worker.NewWorker(cam)
	go timelapseWorker.Record()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		latestFile := getLatestFile(cfg.OutputDir)
		return c.SendString(fmt.Sprintf("Hello, World! %s", latestFile))
	})

	err := app.Listen(":80")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
