package core

import (
	"log"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/hand"
)

type jsonMsg struct {
	ID        int         `json:"id"`
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
	ID        int
	Name      string
	Chips     int
	Hand      []hand.Card
	Bet       int
	CallCheck string
	Active    bool
	AllIn     bool
	Fold      bool
	Leave     bool
}

func NewPlayer(IDPlayer int, name string, chips int) *Player {
	player := &Player{ID: IDPlayer, Name: name, Chips: chips, Hand: []hand.Card{}}
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
