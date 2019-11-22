package core

import (
	"log"
	"server/game/hand"
)

var IDplayer int32 = 0

type Player struct {
	ID    int32
	Name  string
	Chips int32
	Hand  []hand.Card
	Enemy *Player
}

func NewPlayer(name string, chips int32) *Player {
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
	Score    int32  `json:"score"`
}

type Msg struct {
	Command map[string]*jsonMsg
}

func (p *Player) GetState() *jsonMsg {
	msg := &jsonMsg{
		ID:       p.ID,
		Username: p.Name,
		Score:    p.Chips,
	}
	return msg
}

func (p *Player) GiveUp() {
	log.Print("Player gave up: ", p.Name)
}
