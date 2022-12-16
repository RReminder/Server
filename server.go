package rmessage

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

type Server struct {
	uprader       websocket.FastHTTPUpgrader
	FServer       fasthttp.Server
	Events        *Events
	userCount     atomic.Int64
	users         map[int64]*User
	topics        map[string]*topic
	defaultTopics []*topic
	Router        *router.Router
}

func New() *Server {
	s := &Server{
		uprader: websocket.FastHTTPUpgrader{
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
				return true
			},
		},
		topics:        make(map[string]*topic),
		defaultTopics: make([]*topic, 0),
		FServer:       fasthttp.Server{},
		Events:        newEvents(),
		userCount:     atomic.Int64{},
		users:         make(map[int64]*User),
		Router:        router.New(),
	}
	s.Router.ANY("/rmessage", s.rmessage)
	s.FServer.Handler = s.Router.Handler
	return s
}

func (s *Server) rmessage(ctx *fasthttp.RequestCtx) {
	err := s.uprader.Upgrade(ctx, func(ws *websocket.Conn) {
		defer ws.Close()
		user := &User{
			ID:              s.userCount.Add(1),
			ConnectTime:     time.Now(),
			SubcribedTopics: make([]string, 0),
			conn:            ws,
			Online:          true,
		}
		s.users[user.ID] = user
		defer s.offlineUser(user.ID)
		_, ms, err := ws.ReadMessage()
		if err != nil {
			return
		}
		// Start to auth
		authMessage, err := toMessage(ms)
		if err != nil || authMessage.Type != MessageTypeAuth {
			return
		}
		if !s.Events.OnAuth(user, authMessage) {
			return
		}
		messagePool.Put(authMessage)
		// start to subscribe default topics
		for _, defaultTopic := range s.defaultTopics {
			defaultTopic.subcribers = append(defaultTopic.subcribers, user)
		}
		for {
			_, ms, err := ws.ReadMessage()
			if err != nil {
				return
			}
			func() {
				// start to get processor
				msg, err := toMessage(ms)
				if err != nil {
					return
				}
				defer messagePool.Put(msg)
				s.Events.OnMessage(user, msg)
			}()

		}
	})

	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			log.Println(err)
		}
		return
	}
}
func (s *Server) ListenAndServe(addr string) error {
	return s.FServer.ListenAndServe(addr)
}
func (s *Server) ListenAndServeTLS(addr string, certFile string, keyFile string) error {
	return s.FServer.ListenAndServeTLS(addr, certFile, keyFile)
}
func (s *Server) ListenAndServeTLSEmbed(addr string, cert []byte, key []byte) error {
	return s.FServer.ListenAndServeTLSEmbed(addr, cert, key)
}
func (s *Server) AddTopic(topicName string, defaultTopic bool) error {
	newTopic := &topic{
		subcribers: make([]*User, 0),
	}
	if defaultTopic {
		for _, user := range s.users {
			newTopic.subcribers = append(newTopic.subcribers, user)
		}
		s.defaultTopics = append(s.defaultTopics, newTopic)
	}
	s.topics[topicName] = newTopic
	return nil
}
func (s *Server) offlineUser(userID int64) {
	// delete it from the subscribe list
	s.users[userID].Online = false
}
