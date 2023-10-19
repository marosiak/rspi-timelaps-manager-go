package api

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	. "github.com/macrosiak/rspi-timelaps-manager-go/system_stats"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"time"
)

type Api struct {
	cfg            *config.Config
	systemStatsSrv *SystemStatsService
}

func NewApi(app *fiber.App, systemStatsSrv *SystemStatsService) *Api {
	api := &Api{cfg: config.New(), systemStatsSrv: systemStatsSrv}

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Static("/", api.cfg.WebInterfaceFilesPath)

	app.Get("/ws/", websocket.New(api.WebsocketHandler))
	return api
}

type WebsocketResponse struct {
	Stats            *StatsResponse `json:"stats"`
	LastPhotoTakenAt *int64         `json:"lastPhotoTakenAt"`
}

func (a Api) getLastPhotoTakenAt() (*time.Time, error) {
	var latestTime time.Time

	files, err := os.ReadDir(a.cfg.OutputDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileInfo, err := os.Stat(filepath.Join(a.cfg.OutputDir, file.Name()))
		if err != nil {
			return nil, err
		}

		fileTime := fileInfo.ModTime()
		if fileTime.After(latestTime) {
			latestTime = fileTime
		}
	}

	return &latestTime, nil
}

func (a Api) WebsocketHandler(c *websocket.Conn) {
	var (
		mt int
		//msg []byte
		err error
	)
	for {
		if mt, _, err = c.ReadMessage(); err != nil {
			log.Err(err).Msg("read message")
			break
		}

		systemInfo, err := a.systemStatsSrv.GetSystemUsageInfo()
		if err != nil {
			log.Err(err).Msg("get system info")
			continue
		}

		lastPhotoTakenAt, err := a.getLastPhotoTakenAt()
		if err != nil {
			log.Err(err).Msg("get last photo taken at")
			continue
		}

		lastPhotoTakenAtTimestamp := lastPhotoTakenAt.Unix()
		response := WebsocketResponse{
			Stats:            systemInfo,
			LastPhotoTakenAt: &lastPhotoTakenAtTimestamp,
		}

		respJson, err := json.Marshal(response)
		if err != nil {
			log.Err(err).Msg("json marshal")
			break
		}

		if err = c.WriteMessage(mt, respJson); err != nil {
			log.Err(err).Msg("write message")
			break
		}
	}
}
