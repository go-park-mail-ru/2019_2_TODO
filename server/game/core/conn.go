package core

import (
	"log"

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
		log.Print("Command: '", string(command), "' received by player: ", pc.Player.Name)
		if string(command) == "ready" {
			pc.room.RoomReadyCounter++
			// update all conn
			pc.room.UpdateAll <- true
		}
	}
	pc.room.Leave <- pc
	pc.ws.Close()
}

func (pc *playerConn) sendState(command string) {
	go func() {
		msgState := pc.GetState()
		var cmd = make(map[string]*jsonMsg)
		cmd[command] = msgState
		msg := &Msg{
			Command: cmd,
		}
		mutex.Lock()
		err := pc.ws.WriteJSON(msg)
		mutex.Unlock()
		if err != nil {
			pc.room.Leave <- pc
			pc.ws.Close()
		}
	}()
}

func (pc *playerConn) sendNewPlayer(player *playerConn, command string) {
	go func() {
		msgState := player.GetState()
		var cmd = make(map[string]*jsonMsg)
		cmd[command] = msgState
		msg := &Msg{
			Command: cmd,
		}
		mutex.Lock()
		err := pc.ws.WriteJSON(msg)
		mutex.Unlock()
		if err != nil {
			pc.room.Leave <- pc
			pc.ws.Close()
		}
	}()
}

func (pc *playerConn) sendStartGame() {
	go func() {
		mutex.Lock()
		err := pc.ws.WriteJSON(`{"Command":"startGame"}`)
		mutex.Unlock()
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
