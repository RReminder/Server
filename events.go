package rmessage

type AuthFunc func(user *User, message *Message) bool
type MessageFunc func(user *User, message *Message)
type Events struct {
	OnAuth    AuthFunc
	OnMessage MessageFunc
}

func newEvents() *Events {
	return &Events{
		OnAuth: func(user *User, message *Message) bool {
			return true
		},
		OnMessage: func(user *User, message *Message) {},
	}
}
