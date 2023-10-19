package api

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	. "github.com/macrosiak/rspi-timelaps-manager-go/system_stats"
	"github.com/rs/zerolog/log"
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

		respJson, err := json.Marshal(systemInfo)
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
