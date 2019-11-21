package core

import (
	"fmt"
	"log"
	"server/game/hand"
)

type Player struct {
	Name  string
	Chips int32
	Hand  []hand.Card
	Enemy *Player
}

func NewPlayer(name string, chips int32) *Player {
	player := &Player{Name: name, Chips: chips}
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

func (p *Player) GetState() string {
	return "Game state for Player: " + p.Name +
		"\nGame chips: " + fmt.Sprint(p.Chips) +
		"\nHand: " + fmt.Sprint(p.Hand)
}

func (p *Player) GiveUp() {
	log.Print("Player gave up: ", p.Name)
}
