package core

import (
	"github.com/gorilla/websocket"
)

type playerConn struct {
	ws *websocket.Conn
	*Player
	room *Room
}

// Receive msg from ws in goroutine
func (pc *playerConn) receiver() {
	for {
		_, command, err := pc.ws.ReadMessage()
		if err != nil {
			break
		}
		// execute a command
		pc.Command(string(command))
		// update all conn
		pc.room.UpdateAll <- true
	}
	pc.room.Leave <- pc
	pc.ws.Close()
}

func (pc *playerConn) sendState() {
	go func() {
		msg := pc.GetState()
		err := pc.ws.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			pc.room.Leave <- pc
			pc.ws.Close()
		}
	}()
}

func NewPlayerConn(ws *websocket.Conn, player *Player, room *Room) *playerConn {
	pc := &playerConn{ws, player, room}
	go pc.receiver()
	return pc
}
