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
	UserConns map[*UserConn]bool

	// Update state for all conn.
	updateAll chan *UserConn

	// Register requests from the connections.
	Join chan *UserConn

	// Unregister requests from connections.
	Leave chan *UserConn
}

// Run the Room in goroutine
func (r *Room) run() {
	for {
		select {
		case c := <-r.Join:
			r.UserConns[c] = true
			c.Online = true
			c.sendStartChat(c)
			for user := range r.UserConns {
				if user != c {
					user.sendStartChat(c)
				}
			}
			// if Room is full - delete from freeRooms
			if len(r.UserConns) == 2 {
				delete(FreeRooms, r.Name)
			}

		case c := <-r.Leave:
			c.Online = false
			c.LeaveRoom()
			delete(r.UserConns, c)
			if len(r.UserConns) == 0 {
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

func (r *Room) updateAllPlayers(conn *UserConn) {
	for c := range r.UserConns {
		c.sendMsgToUsers(conn)
	}
}

func NewRoom(name string) *Room {
	if name == "" {
		name = utils.RandString(16)
	}

	Room := &Room{
		Name:      name,
		UserConns: make(map[*UserConn]bool),
		updateAll: make(chan *UserConn),
		Join:      make(chan *UserConn),
		Leave:     make(chan *UserConn),
	}

	AllRooms[name] = Room
	FreeRooms[name] = Room

	// run Room
	go Room.run()

	RoomsCount += 1

	return Room
}
