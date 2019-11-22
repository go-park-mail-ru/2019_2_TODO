package core

import (
	"log"
	"server/game/hand"
	"sync"
)

var IDplayer int32 = 0
var mutex = &sync.Mutex{}

type Player struct {
	ID    int32
	Name  string
	Chips int
	Hand  []hand.Card
	Bet int
	Enemy *Player
}

func NewPlayer(name string, chips int) *Player {
	player := &Player{ID: IDplayer, Name: name, Chips: chips}
	IDplayer++
	return player
}

func PairPlayers(p1 *Player, p2 *Player) {
	p1.Enemy, p2.Enemy = p2, p1
}

func (p *Player) Command(command string) {
	log.Print("Command: '", command, "' received by player: ", p.Name)
	if command == "Fold" {

	} else if command == "Call" {

	} else if command == "Raise" {
		p.Chips -= 100
	}
}

type jsonMsg struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Score    int  `json:"score"`
}

type Msg struct {
	Command map[string]*jsonMsg
}

func (p *Player) GetState() *jsonMsg {
	mutex.Lock()
	msg := &jsonMsg{
		ID:       p.ID,
		Username: p.Name,
		Score:    p.Chips,
	}
	mutex.Unlock()
	return msg
}

func (p *Player) GiveUp() {
	log.Print("Player gave up: ", p.Name)
}
