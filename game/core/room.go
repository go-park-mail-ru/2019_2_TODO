package core

import (
	"log"
	"strconv"
	"sync"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"
	repository "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/repositoryLeaders"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/hand"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/utils"
)

var AllRooms = make(map[string]*Room)
var FreeRooms = make(map[string]*Room)
var RoomsCount int

// func init() {
// 	for i := 0; i < 4; i++ {
// 		NewRoom("", 2, false, "", 20)
// 	}
// }

type Room struct {
	Name             string
	RoomReadyCounter int32
	RoomStartRound   bool
	RoomMinBet       int
	Command          string
	Game             *Game

	PlayersInRoom int
	Private       bool
	Password      string

	mu sync.Mutex

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
			if len(r.PlayerConns) == r.PlayersInRoom {
				delete(FreeRooms, r.Name)
			}

		case c := <-r.Leave:
			c.GiveUp()
			userData := &leaderBoardModel.UserLeaderBoard{
				ID:       int64(c.Player.ID),
				Username: c.Player.Name,
				Points:   strconv.Itoa(c.Player.Chips + c.Player.Bet),
			}
			newConn := repository.NewUserMemoryRepository()
			_, err := newConn.UpdateLeader(userData)
			if err != nil {
				log.Println(err)
			}
			r.updateAllPlayersExceptYou(c, "removePlayer")
			c.sendState("removePlayer")
			delete(r.PlayerConns, c)
			if len(r.PlayerConns) == 0 {
				goto Exit
			}
		case c := <-r.UpdateAll:
			if r.RoomStartRound {
				if r.Command == "endFoldGame" {
					r.updateAllPlayers(c, "turnOffPlayer")
				}
				r.updateAllPlayers(c, r.Command)
				for conn := range r.PlayerConns {
					r.updateAllPlayers(conn, "updatePlayerScore")
				}
				if r.Command == "endFoldGame" {
					r.mu.Lock()
					r.RoomReadyCounter = 0
					r.RoomStartRound = false
					r.mu.Unlock()
					goto EndFoldGame
				}
			SkipPlayer:
				if r.Game.AllInCounter == len(r.PlayerConns)-1 {
					for conn := range r.PlayerConns {
						r.updateAllPlayers(conn, "showPlayerCards")
					}
				}
				r.Game.PlayerCounterChange()
				if r.Game.PlayerCounter == r.Game.PositionToNextStage {
					r.Game.StageCounterChange()
					r.setBank()
					r.updateAllPlayersBank("setBank")
					if r.Game.StageCounter == 1 {
						r.updateTableCards(c, "showTableCards", 3)
					} else if r.Game.StageCounter == 2 {
						r.updateTableCards(c, "showTableCards", 4)
					} else if r.Game.StageCounter == 3 {
						r.updateTableCards(c, "showTableCards", 5)
					} else if r.Game.StageCounter == 4 {
						winnerHand := r.endGame()
						r.updateAllPlayersBank("setBank")
						for conn := range r.PlayerConns {
							r.updateAllPlayers(conn, "showPlayerCards")
						}
						for conn := range r.PlayerConns {
							conn.sendWinnerHand(winnerHand, "showWinnerCards")
						}
						r.mu.Lock()
						r.RoomReadyCounter = 0
						r.RoomStartRound = false
						r.mu.Unlock()
						for conn := range r.PlayerConns {
							r.updateAllPlayers(conn, "updatePlayerScore")
						}
						goto EndFoldGame
					}
					for conn := range r.PlayerConns {
						r.updateAllPlayers(conn, "updatePlayerScore")
					}
				}
				if r.Game.Players[r.Game.PlayerCounter].Bet < r.Game.MaxBet {
					r.Game.Players[r.Game.PlayerCounter].CallCheck = "call"
				} else {
					r.Game.Players[r.Game.PlayerCounter].CallCheck = "check"
				}
				if r.Game.Players[r.Game.PlayerCounter].Player.AllIn ||
					r.Game.Players[r.Game.PlayerCounter].Player.Fold ||
					(r.Game.AllInCounter == len(r.PlayerConns)-1) {
					goto SkipPlayer
				}
				r.updateAllPlayers(r.Game.Players[r.Game.PlayerCounter], "enablePlayer")
			}
		EndFoldGame:
			r.mu.Lock()
			allReady := (r.RoomReadyCounter == int32(r.PlayersInRoom))
			started := r.RoomStartRound
			r.mu.Unlock()
			if allReady && !started {
				for conn := range r.PlayerConns {
					conn.AllIn = false
					conn.Fold = false
					if conn.Chips <= 0 {
						conn.GiveUp()
						userData := &leaderBoardModel.UserLeaderBoard{
							ID:       int64(c.Player.ID),
							Username: conn.Player.Name,
							Points:   strconv.Itoa(conn.Player.Chips + conn.Player.Bet),
						}
						newConn := repository.NewUserMemoryRepository()
						_, err := newConn.UpdateLeader(userData)
						if err != nil {
							log.Println(err)
						}
						r.updateAllPlayersExceptYou(conn, "removePlayer")
						c.sendState("removePlayer")
						delete(r.PlayerConns, conn)
					}
				}
				if len(r.PlayerConns) == 0 {
					goto Exit
				}
				players := []*playerConn{}
				for player := range r.PlayerConns {
					player.Active = true
					players = append(players, player)
					player.sendState("startGame")
				}
				r.Game = &Game{
					Players:       players,
					TableCards:    []hand.Card{},
					Bank:          0,
					Dealer:        0,
					MinBet:        r.RoomMinBet,
					PlayerCounter: 0,
					StageCounter:  0,
				}
				r.Game.StartGame()
				r.updateAllPlayersBank("setBank")
				r.Game.MaxBet = r.Game.MinBet * 2
				r.mu.Lock()
				r.RoomStartRound = true
				r.mu.Unlock()
				if r.Game.Players[r.Game.PlayerCounter].Bet < r.Game.MaxBet {
					r.Game.Players[r.Game.PlayerCounter].CallCheck = "call"
				} else {
					r.Game.Players[r.Game.PlayerCounter].CallCheck = "check"
				}
				r.updateAllPlayers(r.Game.Players[r.Game.PlayerCounter], "enablePlayer")
			}
		}
	}

Exit:

	// delete room
	delete(AllRooms, r.Name)
	delete(FreeRooms, r.Name)
	r.mu.Lock()
	RoomsCount -= 1
	r.mu.Unlock()
	log.Print("Room closed:", r.Name)
}

func (r *Room) endGame() []hand.Card {
	var bestHand = []hand.Card{}
	var bestRank int32 = 7463
	var player *playerConn
	for c := range r.PlayerConns {
		log.Println(c.Player.ID)
		log.Println(c.Player.Hand)
		currentHand := c.Player.Hand
		currentHand = append(currentHand, r.Game.TableCards...)
		rankCurrentHand := hand.Evaluate(currentHand)
		if rankCurrentHand < bestRank {
			bestRank = rankCurrentHand
			bestHand = currentHand
			player = c
		}
	}
	player.Player.Chips += r.Game.Bank
	r.Game.Bank = 0
	return bestHand
}

func (r *Room) updateAfterEndGame(winnerHand []hand.Card, command string) {
	for c := range r.PlayerConns {
		c.sendWinnerHand(winnerHand, command)
	}
}

func (r *Room) setBank() {
	for c := range r.PlayerConns {
		r.Game.Bank += c.Player.Bet
		c.Player.Bet = 0
	}
}

func (r *Room) updateAllPlayersBank(command string) {
	for c := range r.PlayerConns {
		c.sendBankState(command)
	}
}

func (r *Room) updateAllPlayersExceptYou(conn *playerConn, command string) {
	for c := range r.PlayerConns {
		if conn != c {
			c.sendNewPlayer(conn, command)
		}
	}
}

func (r *Room) updateTableCards(conn *playerConn, command string, numberCards int) {
	for c := range r.PlayerConns {
		c.sendTableCards(command, numberCards)
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

func NewRoom(name string, playersInRoom int, private bool, password string, minBet int) *Room {
	if name == "" {
		name = utils.RandString(16)
	}

	var mu sync.Mutex

	room := &Room{
		Name:             name,
		RoomReadyCounter: 0,
		Game:             &Game{},
		RoomMinBet:       minBet,
		PlayersInRoom:    playersInRoom,
		Private:          private,
		Password:         password,
		PlayerConns:      make(map[*playerConn]bool),
		UpdateAll:        make(chan *playerConn),
		Join:             make(chan *playerConn),
		Leave:            make(chan *playerConn),
	}

	AllRooms[name] = room
	FreeRooms[name] = room

	// run room
	go room.run()

	mu.Lock()
	RoomsCount += 1
	mu.Unlock()

	return room
}
