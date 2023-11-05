package api

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/rs/zerolog/log"
	"strings"
)

type Topic string

func (t Topic) ToUpper() Topic {
	return Topic(strings.ToUpper(string(t)))
}

const (
	StatisticsTopic Topic = "STATISTICS"
	PhotosTopic     Topic = "PHOTOS"
)

type TopicsWhitelist []Topic

func (w TopicsWhitelist) Contains(topic Topic) bool {
	for _, t := range w {
		if t == topic {
			return true
		}
	}
	return false
}

var WhitelistedTopics = TopicsWhitelist{StatisticsTopic, PhotosTopic}

type Connection struct {
	Conn        *websocket.Conn
	MessageType int // the type of message for websocket
}

type PubSub struct {
	Subscribers map[Topic][]*Connection
}

func NewPubSub() *PubSub {
	return &PubSub{
		Subscribers: make(map[Topic][]*Connection),
	}
}

var TopicNotWhitelistedErr = fmt.Errorf("topic not whitelisted")

func (p *PubSub) Subscribe(c *websocket.Conn, messageType int, topic Topic) error {
	topic = topic.ToUpper()
	if !WhitelistedTopics.Contains(topic) {
		return TopicNotWhitelistedErr
	}
	p.Subscribers[topic] = append(p.Subscribers[topic], &Connection{Conn: c, MessageType: messageType})
	return nil
}

func (p *PubSub) Unsubscribe(c *websocket.Conn, topic Topic) {
	topic = topic.ToUpper()
	for i, subscriber := range p.Subscribers[topic] {
		if subscriber.Conn == c {
			p.Subscribers[topic] = append(p.Subscribers[topic][:i], p.Subscribers[topic][i+1:]...)
			break
		}
	}
}

func (p *PubSub) UnsubscribeFromAll(c *websocket.Conn) {
	for topic := range p.Subscribers {
		p.Unsubscribe(c, topic)
	}
}

func (p *PubSub) Publish(topic Topic, message []byte) {
	if len(p.Subscribers[topic]) == 0 {
		return
	}

	log.Debug().Msgf("Publishing to topic %s: %s", topic, string(message))
	for _, subscriber := range p.Subscribers[topic] {
		if err := subscriber.Conn.WriteMessage(subscriber.MessageType, message); err != nil {
			p.Unsubscribe(subscriber.Conn, topic) // unsubscribe if error
		}
	}
}

func (p *PubSub) PublishJson(topic Topic, message interface{}) error {
	by, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	p.Publish(topic, by)
	return nil
}
