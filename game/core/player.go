package core

import (
	"game/hand"
	"log"
)

var IDplayer int32 = 0

type jsonMsg struct {
	ID        int32       `json:"id"`
	Username  string      `json:"username"`
	Bet       int         `json:"bet"`
	Score     int         `json:"score"`
	Hand      []hand.Card `json:"hand"`
	CallCheck string      `json:"callCheck"`
}

type Msg struct {
	Command map[string]*jsonMsg
}

type Player struct {
	ID        int32
	Name      string
	Chips     int
	Hand      []hand.Card
	Bet       int
	CallCheck string
	Active    bool
}

func NewPlayer(name string, chips int) *Player {
	player := &Player{ID: IDplayer, Name: name, Chips: chips, Hand: []hand.Card{}}
	IDplayer++
	return player
}

func (p *Player) GetState() *jsonMsg {
	msg := &jsonMsg{
		ID:        p.ID,
		Username:  p.Name,
		Bet:       p.Bet,
		Score:     p.Chips,
		Hand:      p.Hand,
		CallCheck: p.CallCheck,
	}
	return msg
}

func (p *Player) GiveUp() {
	log.Print("Player gave up: ", p.Name)
}
