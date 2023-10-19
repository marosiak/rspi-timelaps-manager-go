package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/macrosiak/rspi-timelaps-manager-go/api"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/macrosiak/rspi-timelaps-manager-go/system_stats"
	"github.com/macrosiak/rspi-timelaps-manager-go/views"
	"github.com/macrosiak/rspi-timelaps-manager-go/worker"
	"github.com/rs/zerolog"
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
	if cfg.Development {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var cam camera.Camera
	if cfg.Development {
		cam = camera.NewFakeCamera()
	} else {
		cam = camera.NewLibCamera(&camera.CameraSettings{})
	}

	timelapseWorker := worker.NewWorker(cam, cfg)
	go timelapseWorker.Run()

	engine := html.NewFileSystem(http.FS(views.GetViewsFileSystem()), ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	//app.Static("/", "./web_client")
	systemStatsSrv := system_stats.NewSystemStats()
	if cfg.WebInterface {
		_ = api.NewApi(app, systemStatsSrv)
	}

	err := app.Listen(":80")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
