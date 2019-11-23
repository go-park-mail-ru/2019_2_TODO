package core

import "log"

var IDUser int32 = 0

type Message struct {
	Autor string `json:"author"`
	Body  string `json:"body"`
}

type User struct {
	ID     int32
	Msg    *Message
	Admin  bool
	Online bool
}

func NewUser(name string, admin bool) *User {
	user := &User{
		ID: IDUser,
		Msg: &Message{
			Autor: name,
			Body:  "",
		},
		Admin:  admin,
		Online: false,
	}
	return user
}

func (u *User) UserMessage(msg string) {
	log.Print("Message: '", msg, "' received by user: ", u.Msg.Autor)
	u.Msg.Body = msg
}

func (u *User) UserGetMessage() *Message {
	msg := &Message{
		Autor: u.Msg.Autor,
		Body:  u.Msg.Body,
	}
	return msg
}

func (u *User) LeaveRoom() {
	log.Print("User left: ", u.Msg.Autor)
}

