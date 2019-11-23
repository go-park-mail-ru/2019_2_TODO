package core

import (
	"sync"

	"github.com/gorilla/websocket"
)

var mutex = &sync.Mutex{}

type userConn struct {
	ws *websocket.Conn
	*User
	room *Room
}

// Receive msg from ws in goroutine
func (uc *userConn) receiver() {
	for {
		_, command, err := uc.ws.ReadMessage()
		if err != nil {
			break
		}
		// execute a command
		uc.UserMessage(string(command))
		// update all conn
		uc.room.updateAll <- uc
	}
	uc.room.Leave <- uc
	uc.ws.Close()
}

func (uc *userConn) sendMsgToUsers(user *userConn) {
	go func() {
		msg := user.UserGetMessage()
		mutex.Lock()
		err := uc.ws.WriteJSON(msg)
		mutex.Unlock()
		if err != nil {
			uc.room.Leave <- uc
			uc.ws.Close()
		}
	}()
}

func (uc *userConn) sendStartChat() {
	msg := &Message{
		Autor: uc.Msg.Autor,
		Body:  "Joined Room",
	}
	err := uc.ws.WriteJSON(msg)
	if err != nil {
		uc.room.Leave <- uc
		uc.ws.Close()
	}
}

func NewUserConn(ws *websocket.Conn, user *User, room *Room) *userConn {
	uc := &userConn{ws, user, room}
	go uc.receiver()
	return uc
}
