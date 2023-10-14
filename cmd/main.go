package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/macrosiak/rspi-timelaps-manager-go/views"
	"github.com/macrosiak/rspi-timelaps-manager-go/worker"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"path/filepath"
)

func getLatestFile(dir string) string {
	files, _ := os.ReadDir(dir)
	var newestFile string
	var newestTime int64 = 0
	for _, f := range files {
		fi, err := os.Stat(filepath.Join(dir, f.Name()))
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
	//log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	cfg := config.New()

	var cam camera.Camera
	if cfg.Development {
		cam = camera.NewFakeCamera()
	} else {
		settings := &camera.CameraSettings{
			AutoFocusRange: camera.AutoFocusNormal,
		}
		cam = camera.NewLibCamera(settings)
	}

	timelapseWorker := worker.NewWorker(cam, cfg)
	go timelapseWorker.Run()

	engine := html.NewFileSystem(http.FS(views.GetViewsFileSystem()), ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Get("/latest", func(c *fiber.Ctx) error {
		latestFile := getLatestFile(cfg.OutputDir)
		return c.Redirect("/" + latestFile)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		latestFile := getLatestFile(cfg.OutputDir)
		return c.Render("index", fiber.Map{
			"LatestImage": latestFile,
		})
	})
	app.Use(filesystem.New(filesystem.Config{
		Root:   http.Dir(cfg.OutputDir),
		Browse: true,
		MaxAge: 3600,
	}))

	err := app.Listen(":80")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
