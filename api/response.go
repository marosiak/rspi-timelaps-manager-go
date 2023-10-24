package api

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	. "github.com/macrosiak/rspi-timelaps-manager-go/system_stats"
	"github.com/rs/zerolog/log"
)

type WebsocketStatsResponse struct {
	Stats            *StatsResponse `json:"stats"`
	LastPhotoTakenAt int64          `json:"lastPhotoTakenAt"`
}

type WebsocketError string

const (
	WebsocketNotAuthorisedError WebsocketError = "NOT_AUTHORISED"
)

type WebsocketErrorResponse struct {
	Error WebsocketError `json:"error"`
}

func sendStruct(c *websocket.Conn, mt int, theStruct interface{}) {
	respJson, err := json.Marshal(theStruct)
	if err != nil {
		log.Err(err).Msg("json marshal")
	}

	if err = c.WriteMessage(mt, respJson); err != nil {
		log.Err(err).Msg("write message")
	}
}

func SendError(c *websocket.Conn, mt int, websocketError WebsocketError) {
	response := WebsocketErrorResponse{
		Error: websocketError,
	}

	sendStruct(c, mt, response)
}

type Action string

const (
	ActionRemoveAllImages = "REMOVE_ALL_IMAGES"
	ActionAuth            = "AUTH"
)

type ActionPayload struct {
	Action Action `json:"action"`
	Value  string `json:"value"`
}

type ActionStatus string

const (
	ActionStatusSuccess          = "SUCCESS"
	ActionStatusFail             = "FAIL"
	ActionStatusWrongCredentials = "WRONG_CREDENTIALS"
)

type ActionResponse struct {
	Action  Action       `json:"action"`
	Status  ActionStatus `json:"status"`
	Message *string      `json:"message"`
}

func SendStatus(c *websocket.Conn, mt int, action Action, status ActionStatus, message *string) {
	response := ActionResponse{
		Action:  action,
		Status:  status,
		Message: message,
	}

	sendStruct(c, mt, response)
}
