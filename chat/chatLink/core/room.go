package core

import (
	"log"

	"github.com/alehano/wsgame/utils"
)

var AllRooms = make(map[string]*Room)
var FreeRooms = make(map[string]*Room)
var RoomsCount int

type Room struct {
	Name string

	// Registered connections.
	userConns map[*userConn]bool

	// Update state for all conn.
	updateAll chan *userConn

	// Register requests from the connections.
	Join chan *userConn

	// Unregister requests from connections.
	Leave chan *userConn
}

// Run the Room in goroutine
func (r *Room) run() {
	for {
		select {
		case c := <-r.Join:
			r.userConns[c] = true
			c.Online = true
			c.sendStartChat()

			// if Room is full - delete from freeRooms
			if len(r.userConns) == 2 {
				delete(FreeRooms, r.Name)
			}

		case c := <-r.Leave:
			c.Online = false
			c.LeaveRoom()
			delete(r.userConns, c)
			if len(r.userConns) == 0 {
				goto Exit
			}
		case c := <-r.updateAll:
			r.updateAllPlayers(c)
		}
	}

Exit:

	// delete Room
	delete(AllRooms, r.Name)
	delete(FreeRooms, r.Name)
	RoomsCount -= 1
	log.Print("Room closed:", r.Name)
}

func (r *Room) updateAllPlayers(conn *userConn) {
	for c := range r.userConns {
		c.sendMsgToUsers(conn)
	}
}

func NewRoom(name string) *Room {
	if name == "" {
		name = utils.RandString(16)
	}

	Room := &Room{
		Name:      name,
		userConns: make(map[*userConn]bool),
		updateAll: make(chan *userConn),
		Join:      make(chan *userConn),
		Leave:     make(chan *userConn),
	}

	AllRooms[name] = Room
	FreeRooms[name] = Room

	// run Room
	go Room.run()

	RoomsCount += 1

	return Room
}
