package rmessage

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

type Server struct {
	uprader       websocket.FastHTTPUpgrader
	server        fasthttp.Server
	Events        *Events
	userCount     atomic.Int64
	users         map[int64]*User
	topics        map[string]*topic
	defaultTopics []string
}

func New() *Server {
	s := &Server{
		uprader:       websocket.FastHTTPUpgrader{},
		topics:        make(map[string]*topic),
		defaultTopics: make([]string, 0),
		server:        fasthttp.Server{},
	}
	s.server.Handler = func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/rmessage":
			s.rmessage(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}
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
		}
		s.users[user.ID] = user
		_, ms, err := ws.ReadMessage()
		if err != nil {
			return
		}
		// Start to auth
		authMessage, err := toMessage(ms)
		if err != nil {
			return
		}
		s.Events.OnAuth(user, authMessage)
		messagePool.Put(authMessage)
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
	return s.server.ListenAndServe(addr)
}
func (s *Server) ListenAndServeTLS(addr string, certFile string, keyFile string) error {
	return s.server.ListenAndServeTLS(addr, certFile, keyFile)
}
func (s *Server) ListenAndServeTLSEmbed(addr string, cert []byte, key []byte) error {
	return s.server.ListenAndServeTLSEmbed(addr, cert, key)
}
func (s *Server) AddTopic(topicName string, defaultTopic bool) error {
	newTopic := &topic{
		subcribers: make([]*User, 0),
	}
	s.topics[topicName] = newTopic
	if defaultTopic {
		s.defaultTopics = append(s.defaultTopics, topicName)
		for _, user := range s.users {
			newTopic.subcribers = append(newTopic.subcribers, user)
		}
	}
	return nil
}
