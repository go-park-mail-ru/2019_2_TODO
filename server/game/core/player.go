package core

import (
	"log"
	"server/game/hand"
	"strconv"
	"strings"
	"sync"
)

var IDplayer int32 = 0
var mutex = &sync.Mutex{}

type jsonMsg struct {
	ID       int32       `json:"id"`
	Username string      `json:"username"`
	Bet      int         `json:"bet"`
	Score    int         `json:"score"`
	Hand     []hand.Card `json:"hand"`
}

type Msg struct {
	Command map[string]*jsonMsg
}

type Player struct {
	ID    int32
	Name  string
	Chips int
	Hand  []hand.Card
	Bet   int
}

func NewPlayer(name string, chips int) *Player {
	player := &Player{ID: IDplayer, Name: name, Chips: chips, Hand: []hand.Card{}}
	IDplayer++
	return player
}

func (p *Player) Command(command string) string {
	log.Print("Command: '", command, "' received by player: ", p.Name)
	if command == "fold" {
		p.Hand = []hand.Card{}
		command = "turnOffPlayer"
	} else if command == "check" {
		command = "setCheck"
	} else if command == "call" {
		p.Chips -= MaxBet - p.Bet
		p.Bet = MaxBet
		command = "updatePlayerScore"
	} else {
		raiseCommand := strings.Split(command, " ")
		command = "updatePlayerScore"
		bet, err := strconv.Atoi(raiseCommand[1])
		if err != nil {
			log.Println("error")
		}
		p.Bet = bet
		p.Chips -= bet
		MaxBet = bet
	}
	return command
}

func (p *Player) GetState() *jsonMsg {
	msg := &jsonMsg{
		ID:       p.ID,
		Username: p.Name,
		Bet:      p.Bet,
		Score:    p.Chips,
		Hand:     p.Hand,
	}
	return msg
}

func (p *Player) GiveUp() {
	log.Print("Player gave up: ", p.Name)
}
