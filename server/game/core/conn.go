package core

import (
	"log"
	"server/game/hand"

	"github.com/gorilla/websocket"
)

type playerConn struct {
	ws *websocket.Conn
	*Player
	Room *Room
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
			pc.Room.RoomReadyCounter++
		} else {
			Command = pc.Command(string(command))
		}
		// update all conn
		pc.Room.UpdateAll <- pc
	}
	pc.Room.Leave <- pc
	pc.ws.Close()
}

func (pc *playerConn) sendState(command string) {
	msgState := pc.GetState()
	var cmd = make(map[string]*jsonMsg)
	cmd[command] = msgState
	msg := &Msg{
		Command: cmd,
	}
	err := pc.ws.WriteJSON(msg)
	if err != nil {
		pc.Room.Leave <- pc
		pc.ws.Close()
	}
}

func (pc *playerConn) sendBankState(command string) {
	msgState := &jsonMsg{
		Score: pc.Room.Game.Bank,
	}
	var cmd = make(map[string]*jsonMsg)
	cmd[command] = msgState
	msg := &Msg{
		Command: cmd,
	}
	err := pc.ws.WriteJSON(msg)
	if err != nil {
		pc.Room.Leave <- pc
		pc.ws.Close()
	}
}

func (pc *playerConn) sendNewPlayer(player *playerConn, command string) {
	msgState := player.GetState()
	var cmd = make(map[string]*jsonMsg)
	cmd[command] = msgState
	msg := &Msg{
		Command: cmd,
	}
	err := pc.ws.WriteJSON(msg)
	if err != nil {
		pc.Room.Leave <- pc
		pc.ws.Close()
	}
}

type TableJSON struct {
	Indexes []int       `json:"indexes"`
	Cards   []hand.Card `json:"cards"`
}

type TableCardMsg struct {
	Command map[string]*TableJSON
}

func (pc *playerConn) sendTableCards(command string, numberCards int) {
	indexes := []int{}
	for i := 0; i < numberCards; i++ {
		indexes = append(indexes, i)
	}
	tableJSON := &TableJSON{
		Indexes: indexes,
		Cards:   pc.Room.Game.TableCards[:numberCards],
	}
	var cmd = make(map[string]*TableJSON)
	cmd[command] = tableJSON
	tableCardMsg := &TableCardMsg{
		Command: cmd,
	}
	err := pc.ws.WriteJSON(tableCardMsg)
	if err != nil {
		pc.Room.Leave <- pc
		pc.ws.Close()
	}
}

func (pc *playerConn) sendStartGame() {
	err := pc.ws.WriteJSON(`{"Command":"StartGame"}`)
	if err != nil {
		pc.Room.Leave <- pc
		pc.ws.Close()
	}
}

func NewPlayerConn(ws *websocket.Conn, player *Player, room *Room) *playerConn {
	pc := &playerConn{ws, player, room}
	go pc.receiver()
	return pc
}
