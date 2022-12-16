package rmessage

import "sync"

type topic struct {
	sync.RWMutex
	subcribers []*User
}
