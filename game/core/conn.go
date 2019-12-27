package core

import (
	"log"
	"strconv"
	"strings"

	hand "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/hand"

	"github.com/gorilla/websocket"
)

type playerConn struct {
	ws *websocket.Conn
	*Player
	Room *Room
}

type TableJSON struct {
	Indexes []int       `json:"indexes"`
	Cards   []hand.Card `json:"cards"`
}

type TableCardMsg struct {
	Command map[string]*TableJSON
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
			pc.Room.Command = pc.Command(string(command))
		}
		// update all conn
		pc.Room.UpdateAll <- pc
	}
	pc.Room.Leave <- pc
	pc.ws.Close()
}

func (pc *playerConn) Command(command string) string {
	log.Print("Command: '", command, "' received by player: ", pc.Player.Name)
	if command == "fold" {
		pc.Room.Game.ActivePlayers--
		pc.Player.Hand = []hand.Card{}
		pc.Player.Active = false
		pc.Player.Fold = true
		counterOfActivePlayers := 0
		pc.Room.Game.Bank += pc.Player.Bet
		for c := range pc.Room.PlayerConns {
			if c.Player.Active == true {
				counterOfActivePlayers++
			}
		}
		if counterOfActivePlayers == 1 {
			command = "endFoldGame"
			for c := range pc.Room.PlayerConns {
				if c.Player.Active == true {
					pc.Room.Game.Bank += c.Player.Bet
					c.Player.Chips += pc.Room.Game.Bank
				}
			}
		} else {
			command = "turnOffPlayer"
		}
	} else if command == "check" {
		command = "setCheck"
	} else if command == "call" {
		if pc.Player.Chips < (pc.Room.Game.MaxBet - pc.Player.Bet) {
			pc.Player.Bet += pc.Player.Chips
			pc.Player.Chips = 0
		} else {
			pc.Player.Chips -= pc.Room.Game.MaxBet - pc.Player.Bet
			pc.Player.Bet = pc.Room.Game.MaxBet
		}
		command = "updatePlayerScore"
	} else {
		raiseCommand := strings.Split(command, " ")
		command = "updatePlayerScore"
		bet, err := strconv.Atoi(raiseCommand[1])
		if err != nil {
			log.Println("error")
		}
		pc.Player.Bet += bet
		pc.Player.Chips -= bet
		pc.Room.Game.MaxBet = pc.Player.Bet
		pc.Room.Game.PositionToNextStage = pc.Room.Game.PlayerCounter
	}
	if pc.Player.Chips <= 0 {
		pc.Player.AllIn = true
		pc.Room.Game.AllInCounter++
	}
	return command
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

type jsonMinBet struct {
	MinBet int `json:"minbet"`
}

type jsonMinBetMsg struct {
	Command map[string]*jsonMinBet
}

func (pc *playerConn) sendMinBet(command string, maxBetInGame int) {
	var minBet int
	if (maxBetInGame - pc.Player.Bet) < pc.Room.Game.MinBet {
		minBet = pc.Room.Game.MinBet
	} else {
		minBet = maxBetInGame - pc.Player.Bet
	}
	msgState := &jsonMinBet{
		MinBet: minBet,
	}
	var cmd = make(map[string]*jsonMinBet)
	cmd[command] = msgState
	msg := &jsonMinBetMsg{
		Command: cmd,
	}
	err := pc.ws.WriteJSON(msg)
	if err != nil {
		pc.Room.Leave <- pc
		pc.ws.Close()
	}
}

func (pc *playerConn) sendWinnerHand(winnerHand []hand.Card, command string) {
	msgState := &jsonMsg{
		Hand: winnerHand,
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
