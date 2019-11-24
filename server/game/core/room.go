package core

import (
	"log"

	"server/game/hand"
	"server/game/utils"
)

var AllRooms = make(map[string]*Room)
var FreeRooms = make(map[string]*Room)
var RoomsCount int
var MaxBet int
var Command string

type Room struct {
	Name             string
	RoomReadyCounter int32
	RoomStartGame    bool
	game             *Game

	// Registered connections.
	PlayerConns map[*playerConn]bool

	// Update state for all conn.
	UpdateAll chan *playerConn

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
			r.updateAllPlayersExceptYou(c, "addPlayer")
			r.updateLastPlayer(c, "addPlayer")

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
		case c := <-r.UpdateAll:
			if r.RoomStartGame {
				r.updateAllPlayers(c, Command)
				r.game.PlayerCounterChange()
				r.updateAllPlayers(r.game.Players[r.game.PlayerCounter], "enablePlayer")
			}
			if r.RoomReadyCounter == 2 && !r.RoomStartGame {
				log.Println("All Players are Ready")
				players := []*playerConn{}
				for player := range r.PlayerConns {
					players = append(players, player)
					player.sendState("startGame")
				}
				r.game = &Game{
					Players:       players,
					TableCards:    []hand.Card{},
					Bank:          0,
					Dealer:        0,
					MinBet:        20,
					PlayerCounter: 0,
				}
				r.game.StartGame()
				MaxBet = r.game.MinBet * 2
				r.RoomStartGame = true
				r.updateAllPlayers(r.game.Players[r.game.PlayerCounter], "enablePlayer")
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

func (r *Room) updateAllPlayersExceptYou(conn *playerConn, command string) {
	for c := range r.PlayerConns {
		if conn != c {
			c.sendNewPlayer(conn, command)
		}
	}
}

func (r *Room) updateAllPlayers(conn *playerConn, command string) {
	for c := range r.PlayerConns {
		c.sendNewPlayer(conn, command)
	}
}

func (r *Room) updateLastPlayer(conn *playerConn, command string) {
	for c := range r.PlayerConns {
		if conn != c {
			conn.sendNewPlayer(c, command)
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
		game:             &Game{},
		PlayerConns:      make(map[*playerConn]bool),
		UpdateAll:        make(chan *playerConn),
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
