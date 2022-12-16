package rmessage

import (
	"time"

	"github.com/fasthttp/websocket"
)

type User struct {
	ID              int64
	ConnectTime     time.Time
	conn            *websocket.Conn
	SubcribedTopics []string
	Online          bool
}
