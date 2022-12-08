package rmessage

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fastjson"
)

type Message struct {
	Payload *fastjson.Value `json:"payload"`
	Type    int             `json:"type"`
	Topic   string          `json:"topic"`
}

const (
	MessageTypeAuth      = 0
	MessageTypeBroadcast = 1
	MessageTypeAssigned  = 2
)

var messagePool = sync.Pool{
	New: func() any {
		return Message{}
	},
}

func toMessage(msg []byte) (*Message, error) {
	message := messagePool.Get().(*Message)
	val, err := fastjson.ParseBytes(msg)
	if err != nil {
		return nil, err
	}
	message.Type = val.GetInt("type")
	message.Topic = string(val.GetStringBytes("topic"))
	message.Payload = val.Get("payload")
	return message, nil
}

type rtMessage struct {
	Payload interface{} `json:"payload"`
	Type    int         `json:"type"`
	Topic   string      `json:"topic"`
}

func (s *Server) BroadCast(topic string, payload interface{}) error {
	var rxMessage rtMessage
	rxMessage.Payload = payload
	rxMessage.Topic = topic
	rxMessage.Type = MessageTypeBroadcast
	rtopic, ok := s.topics[topic]
	if !ok {
		return errors.New("unknown topic")
	}
	data, err := json.Marshal(rxMessage)
	if err != nil {
		return err
	}
	for _, user := range rtopic.subcribers {
		user.conn.WriteMessage(websocket.TextMessage, data)
	}
	return nil
}
func (s *Server) SendToUser(topic string, payload interface{}, userID int64) error {
	var rxMessage rtMessage
	rxMessage.Payload = payload
	rxMessage.Topic = topic
	rxMessage.Type = MessageTypeAssigned
	data, err := json.Marshal(rxMessage)
	if err != nil {
		return err
	}
	user, ok := s.users[userID]
	if !ok {
		return errors.New("unknown user")
	}
	return user.conn.WriteMessage(websocket.TextMessage, data)
}
