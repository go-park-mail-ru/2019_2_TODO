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
	Name             string
	RoomReadyCounter int32

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
			c.sendState("addPlayer")
			r.updateAllPlayers(c)
			r.updateLastPlayer(c)

			// if room is full - delete from freeRooms
			if len(r.PlayerConns) == 2 {
				delete(FreeRooms, r.Name)

			}

		case c := <-r.Leave:
			c.GiveUp()
			// r.updateAllPlayers()
			delete(r.PlayerConns, c)
			if len(r.PlayerConns) == 0 {
				goto Exit
			}
		case <-r.UpdateAll:
			if r.RoomReadyCounter == 2 {
				log.Println("Ready")
				players := []*playerConn{}
				for player := range r.PlayerConns {
					players = append(players, player)
				}
				game := &Game{
					Players:       players,
					TableCards:    []hand.Card{},
					Bank:          0,
					Dealer:        0,
					MinBet:        20,
					PlayerCounter: 0,
				}
				game.StartGame()
			}
		}
	}

Exit:

	// delete room
	delete(AllRooms, r.Name)
	delete(FreeRooms, r.Name)
	RoomsCount -= 1
	log.Print("Room closed:", r.Name)
}

func (r *Room) updateAllPlayers(conn *playerConn) {
	for c := range r.PlayerConns {
		log.Println(conn.GetState())
		log.Println(c.GetState())
		if conn != c {
			c.sendNewPlayer(conn)
		}
	}
}

func (r *Room) updateLastPlayer(conn *playerConn) {
	for c := range r.PlayerConns {
		if conn != c {
			conn.sendNewPlayer(c)
		}
	}
}

func NewRoom(name string) *Room {
	if name == "" {
		name = utils.RandString(16)
	}

	room := &Room{
		Name:             name,
		RoomReadyCounter: 0,
		PlayerConns:      make(map[*playerConn]bool),
		UpdateAll:        make(chan bool),
		Join:             make(chan *playerConn),
		Leave:            make(chan *playerConn),
	}

	AllRooms[name] = room
	FreeRooms[name] = room

	// run room
	go room.run()

	RoomsCount += 1

	return room
}
