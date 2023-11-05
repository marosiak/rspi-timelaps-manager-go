package api

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	. "github.com/macrosiak/rspi-timelaps-manager-go/commands"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	. "github.com/macrosiak/rspi-timelaps-manager-go/system_stats"
	"github.com/rs/zerolog/log"
	"time"
)

type Api struct {
	cfg               *config.Config
	systemStatsSrv    *StatisticsService
	connectionsAuthed map[*websocket.Conn]bool
	commandsService   *CommendsService
	pubSub            *PubSub
}

func (a Api) authApiKey(c *websocket.Conn, key string) bool {
	if key == a.cfg.Password {
		a.connectionsAuthed[c] = true
		return true
	}
	return false
}

func (a Api) isUserAuthorised(c *websocket.Conn) bool {
	return a.connectionsAuthed[c]
}

func NewApi(app *fiber.App, systemStatsSrv *StatisticsService, pubSub *PubSub) *Api {
	cfg := config.New()
	api := &Api{cfg: cfg, systemStatsSrv: systemStatsSrv, connectionsAuthed: make(map[*websocket.Conn]bool), pubSub: pubSub, commandsService: NewCommendsService(cfg)}
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Static("/", api.cfg.WebInterfaceFilesPath, fiber.Static{
		CacheDuration: time.Hour * 24,
	})
	app.Static("/photos", api.cfg.OutputDir)
	app.Get("/ws/", websocket.New(api.WebsocketHandler))
	go api.StatisticsWorker()
	return api
}

func (a Api) StatisticsWorker() {
	for {
		stats, err := a.systemStatsSrv.GetStats()
		if err != nil {
			log.Err(err).Msg("get stats")
		}

		err = a.pubSub.PublishJson(StatisticsTopic, stats)
		if err != nil {
			log.Err(err).Msg("publish stats")
		}
		time.Sleep(1 * time.Second)
	}
}

func (a Api) WebsocketHandler(c *websocket.Conn) {
	var (
		mt  int
		msg []byte
		err error
	)
	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			log.Err(err).Msg("read message")
			break
		}

		if len(msg) > 0 {
			actionPayload := ActionPayload{}
			err := json.Unmarshal(msg, &actionPayload)
			if err != nil {
				log.Err(err).Str("msg", string(msg)).Msg("unmarshal")
			}

			if !a.isUserAuthorised(c) && actionPayload.Action != ActionAuth {
				SendError(c, mt, ActionStatusNotAuthorisedError)
				continue
			}

			switch actionPayload.Action {
			case ActionAuth:
				if !a.authApiKey(c, actionPayload.Value) {
					SendStatus(c, mt, ActionAuth, ActionStatusWrongCredentials, nil)
					time.Sleep(5 * time.Second) // Sleep to slow down brute force attacks
					// TODO: implement amount of attempts before being kicked, or make it sleep longer every attempt
					continue
				} else {
					SendStatus(c, mt, ActionAuth, ActionStatusSuccess, nil)
					err := a.pubSub.Subscribe(c, mt, StatisticsTopic)
					if err != nil {
						log.Err(err).Msg("subscribe to stats topic after auth")
					}
					continue
				}
			case ActionRemoveAllImages:
				err := a.commandsService.RemoveAllPhotos()
				if err != nil {
					log.Err(err).Msg("remove all images")
				} else {
					log.Debug().Msg("removed all images")
					SendStatus(c, mt, ActionRemoveAllImages, ActionStatusSuccess, nil)
				}
				continue
			case ActionSubscribe:
				err := a.pubSub.Subscribe(c, mt, Topic(actionPayload.Value))
				if err != nil {
					if errors.Is(err, TopicNotWhitelistedErr) {
						SendError(c, mt, ActionStatusInvalidTopic)
						continue
					}
					SendError(c, mt, ActionStatusUnknownError)
					continue
				}
			}
		} else {
			if !a.isUserAuthorised(c) {
				SendError(c, mt, ActionStatusNotAuthorisedError)
				continue
			}
		}
	}
}
