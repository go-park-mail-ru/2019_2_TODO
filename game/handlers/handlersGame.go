package handlers

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/leaderBoardModel"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/core"
	repository "github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/game/repositoryLeaders"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

const (
	LEADERSSIZE int = 10
)

type HandlersGame struct {
	Usecase *repository.LeadersRepository
}

type PlayerInRoom struct {
	username string `json:"username"`
	avatar   string `json:"avatar"`
}

type RoomsInside struct {
	places       int             `json:"places"`
	actualPlaces int             `json:"actualPlaces"`
	players      []*PlayerInRoom `json:"players"`
}

type JSONRooms struct {
	Rooms map[string]*RoomsInside `json:"rooms"`
}

type JSONLeaders struct {
	Leaders []*leaderBoardModel.UserLeaderBoard `json:"leaders"`
}

var (
	SessManager session.AuthCheckerClient
)

func (h *HandlersGame) GetRooms(ctx echo.Context) error {
	ws, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(ctx.Response(), "Not a websocket handshake", 400)
		return err
	} else if err != nil {
		return err
	}

	go func() {
		for {
			if len(core.FreeRooms) == 2 {
				for i := 0; i < 4; i++ {
					core.NewRoom("")
				}
			}
			var rooms = map[string]*RoomsInside{}
			var playersInRoom = []*PlayerInRoom{}
			for r, room := range core.FreeRooms {
				if len(room.PlayerConns) > 0 {
					for pl := range room.PlayerConns {
						userData, err := h.Usecase.SelectUserByID(int64(pl.ID))
						if err != nil {
							log.Println(err)
							break
						}
						player := &PlayerInRoom{
							username: userData.Username,
							avatar:   userData.Avatar,
						}
						playersInRoom = append(playersInRoom, player)
					}
				}
				rooms[r] = &RoomsInside{
					places:       2,
					actualPlaces: len(room.PlayerConns),
					players:      playersInRoom,
				}
				log.Println(rooms[r])
			}
			log.Println(rooms)
			var jsonRooms = &JSONRooms{
				Rooms: rooms,
			}
			err := ws.WriteJSON(jsonRooms)
			if err != nil {
				ws.Close()
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	return nil
}

func (h *HandlersGame) WsHandler(ctx echo.Context) error {
	ws, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(ctx.Response(), "Not a websocket handshake", 400)
		return err
	} else if err != nil {
		return err
	}

	params, err := url.ParseQuery(ctx.Request().URL.RawQuery)
	if err != nil || !(len(params["id"]) > 0) {
		return ctx.JSON(http.StatusInternalServerError, "Smth wrong with parseQuery")
	}

	userID, err := strconv.Atoi(params["id"][0])
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Smth wrong with cookie")
	}

	user, err := h.Usecase.SelectLeaderByID(int64(userID))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Smth wrong with database")
	}

	playerName := user.Username
	playerStartChips, err := strconv.Atoi(user.Points)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Atoi error")
	}

	var roomName string = "newRoom"
	if len(params["roomName"]) > 0 {
		roomName = params["roomName"][0]
	}

	// Get or create a room
	var room *core.Room
	if roomName != "newRoom" {
		room = core.AllRooms[roomName]
	} else {
		room = core.NewRoom("")
	}

	// Create Player and Conn
	player := core.NewPlayer(userID, playerName, playerStartChips)
	pConn := core.NewPlayerConn(ws, player, room)
	// Join Player to room
	room.Join <- pConn

	log.Printf("Player: %s has joined to room: %s", pConn.Name, room.Name)

	return nil
}

func (h *HandlersGame) LeaderBoardTopHandler(ctx echo.Context) error {
	leaders, err := h.Usecase.ListAllLeaders()
	if err != nil {
		return err
	}

	result := &JSONLeaders{
		Leaders: partitionSort(leaders),
	}

	return ctx.JSON(http.StatusOK, result)
}

func partitionSort(leaders []*leaderBoardModel.UserLeaderBoard) []*leaderBoardModel.UserLeaderBoard {
	var result []*leaderBoardModel.UserLeaderBoard

	var tmp *leaderBoardModel.UserLeaderBoard = &leaderBoardModel.UserLeaderBoard{}
	var maxIndex int

	for i := 0; i < LEADERSSIZE; i++ {
		tmp = leaders[0]
		maxIndex = 0

		for j := 1; j < len(leaders); j++ {
			firstValue, err := strconv.Atoi(leaders[j].Points)
			if err != nil {
				log.Println(err)
			}
			secondValue, err := strconv.Atoi(tmp.Points)
			if err != nil {
				log.Println(err)
			}
			if firstValue > secondValue {
				tmp = leaders[j]
				maxIndex = j
			}
		}

		result = append(result, tmp)

		leaders[maxIndex] = leaders[len(leaders)-1]
		leaders[len(leaders)-1] = nil
		leaders = leaders[:len(leaders)-1]

	}
	return result
}
