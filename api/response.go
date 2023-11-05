package api

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	. "github.com/macrosiak/rspi-timelaps-manager-go/system_stats"
	"github.com/rs/zerolog/log"
)

type WebsocketStatsResponse struct {
	Stats *StatsResponse `json:"stats"`
}

type WebsocketErrorResponse struct {
	Error ActionStatus `json:"error"`
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

func SendError(c *websocket.Conn, mt int, websocketError ActionStatus) {
	response := WebsocketErrorResponse{
		Error: websocketError,
	}

	sendStruct(c, mt, response)
}

type Action string

const (
	ActionRemoveAllImages = "REMOVE_ALL_IMAGES"
	ActionAuth            = "AUTH"
	ActionSubscribe       = "SUBSCRIBE"
	ActionUnsubscribe     = "UNSUBSCRIBE"
)

type ActionPayload struct {
	Action Action `json:"action"`
	Value  string `json:"value"`
}

type ActionStatus string

const (
	ActionStatusSuccess            ActionStatus = "SUCCESS"
	ActionStatusUnknownError       ActionStatus = "UNKNOWN_ERROR"
	ActionStatusWrongCredentials   ActionStatus = "WRONG_CREDENTIALS"
	ActionStatusInvalidTopic       ActionStatus = "INVALID_TOPIC"
	ActionStatusNotAuthorisedError ActionStatus = "NOT_AUTHORISED"
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

type PhotoResponse struct {
	Photo     string `json:"photo"`
	CreatedAt int64  `json:"createdAt"`
}
