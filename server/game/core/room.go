package core

import (
	"log"

	"server/game/hand"
	"server/game/utils"
)

var AllRooms = make(map[string]*Room)
var FreeRooms = make(map[string]*Room)
var RoomsCount int

type Room struct {
	Name string

	// Registered connections.
	PlayerConns map[*playerConn]bool

	// Update state for all conn.
	UpdateAll chan bool

	// Register requests from the connections.
	Join chan *playerConn

	// Unregister requests from connections.
	Leave chan *playerConn
}

// Run the room in goroutine
func (r *Room) run() {
	for {
		select {
		case c := <-r.Join:
			r.PlayerConns[c] = true
			r.updateAllPlayers()

			// if room is full - delete from freeRooms
			if len(r.PlayerConns) == 2 {
				delete(FreeRooms, r.Name)
				// pair players
				deck := hand.NewDeck()
				var p []*Player
				for k, _ := range r.PlayerConns {
					p = append(p, k.Player)
				}
				p[0].Hand = deck.Draw(2)
				p[1].Hand = deck.Draw(2)
				PairPlayers(p[0], p[1])
			}

		case c := <-r.Leave:
			c.GiveUp()
			r.updateAllPlayers()
			delete(r.PlayerConns, c)
			if len(r.PlayerConns) == 0 {
				goto Exit
			}
		case <-r.UpdateAll:
			r.updateAllPlayers()
		}
	}

Exit:

	// delete room
	delete(AllRooms, r.Name)
	delete(FreeRooms, r.Name)
	RoomsCount -= 1
	log.Print("Room closed:", r.Name)
}

func (r *Room) updateAllPlayers() {
	for c := range r.PlayerConns {
		c.sendState()
	}
}

func NewRoom(name string) *Room {
	if name == "" {
		name = utils.RandString(16)
	}

	room := &Room{
		Name:        name,
		PlayerConns: make(map[*playerConn]bool),
		UpdateAll:   make(chan bool),
		Join:        make(chan *playerConn),
		Leave:       make(chan *playerConn),
	}

	AllRooms[name] = room
	FreeRooms[name] = room

	// run room
	go room.run()

	RoomsCount += 1

	return room
}